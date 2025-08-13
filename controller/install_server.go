/*
Package controller 包含安装配置服务器和路由处理器

本文件实现了 Dash 博客系统的安装配置接口，提供以下功能：
1. 启动独立的安装配置 HTTP 服务器
2. 验证 MySQL 数据库连接和 Redis 连接
3. 将安装参数写入 conf/install.yaml 并更新 conf/config.yaml
4. 支持通过环境变量或 Docker 容器化部署

注意：安装完成时，install.yaml 中的 status 字段需从 installing 更新为 installed
*/
package controller

import (
	"context"
	"dash/model/dto"
	"dash/model/param"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	// 移除未使用的标准库导入
	// "database/sql"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var quit = make(chan string, 1)

type InstallServer struct {
	HttpServer *http.Server
}

func NewInstallServer() *InstallServer {
	// 创建Gin路由引擎
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	// cfg := cors.Config{ // 创建 CORS 配置
	// 	AllowMethods: []string{"PUT", "PATCH", "GET", "DELETE", "POST", "OPTIONS"}, // 允许的方法
	// 	AllowHeaders: []string{
	// 		"Origin",
	// 		"Authorization",
	// 		"Content-Type",
	// 		"Accept",
	// 		"X-Requested-With",
	// 		"Access-Control-Request-Method",
	// 		"Access-Control-Request-Headers",
	// 	},
	// 	AllowCredentials: true, // 允许携带凭证（如 Cookie）
	// 	ExposeHeaders: []string{
	// 		"Content-Length",
	// 		"Content-Type",
	// 		"Content-Disposition",
	// 		"Access-Control-Allow-Origin",
	// 	},
	// }
	// // 允许任意 localhost 和 127.0.0.1 端口来源，便于前端在不同端口启动（如 5137/5173 等）
	// cfg.AllowOriginFunc = func(origin string) bool { // 自定义来源校验函数
	// 	return strings.HasPrefix(origin, "http://localhost:") || // 放行 localhost 任意端口
	// 		strings.HasPrefix(origin, "http://127.0.0.1:") // 放行 127.0.0.1 任意端口
	// }
	// router.Use(cors.New(cfg)) // 注册 CORS 中间件

	staticRouter := router.Group("/")
	{
		staticRouter.GET("", func(c *gin.Context) {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
			c.File("resource/static/install.html")
		})
		staticRouter.StaticFS("/assets", gin.Dir("resource/static/assets", false)) // 挂载静态资源目录（JS/CSS/图片等）
	}

	router.POST("/api/install", func(ctx *gin.Context) {
		installParams := &param.Install{}
		err := ctx.ShouldBindJSON(installParams)
		if err != nil {
			ctx.JSON(200, gin.H{
				"status":  http.StatusBadRequest,
				"message": "参数无效",
				"data":    nil,
			})
			return
		}

		// 验证数据库连接
		dbType := strings.ToLower(installParams.Database.Type) // 兼容大小写
		switch dbType {
		case "mysql":
			if err := validateMySQLConnection(installParams.Database); err != nil {
				ctx.JSON(200, gin.H{
					"status":  http.StatusBadRequest,
					"message": fmt.Sprintf("MySQL 连接失败: %s", err.Error()),
					"data":    nil,
				})
				return
			}
		case "sqlite", "sqlite3":
			// sqlite 不需要连通性验证
		default:
			ctx.JSON(200, gin.H{
				"status":  http.StatusBadRequest,
				"message": "不支持的数据库类型，仅支持 mysql 或 sqlite3",
				"data":    nil,
			})
			return
		}

		// 验证 Redis 连接
		if err := validateRedisConnection(installParams.Redis); err != nil {
			ctx.JSON(200, gin.H{
				"status":  http.StatusBadRequest,
				"message": fmt.Sprintf("Redis 连接失败: %s", err.Error()),
				"data":    nil,
			})
			return
		}

		// 写入 install.yml 文件
		if err := writeInstallYAML(installParams); err != nil {
			ctx.JSON(200, gin.H{
				"status":  http.StatusInternalServerError,
				"message": fmt.Sprintf("写入安装配置失败: %s", err.Error()),
				"data":    nil,
			})
			return
		}

		// 更新 config.yaml 配置
		if err := updateConfigYAML(installParams); err != nil {
			ctx.JSON(200, gin.H{
				"status":  http.StatusInternalServerError,
				"message": fmt.Sprintf("更新配置文件失败: %s", err.Error()),
				"data":    nil,
			})
			return
		}

		ctx.JSON(200, gin.H{
			"status":  http.StatusOK,
			"message": "安装配置保存成功",
			"data":    nil,
		})
		quit <- "finish"
	})

	router.NoRoute(func(ctx *gin.Context) { // 注册未匹配路由的兜底处理
		path := ctx.Request.URL.Path         // 获取请求路径
		if strings.HasPrefix(path, "/api") { // 如果是 API 路径
			ctx.JSON(http.StatusNotFound, dto.BaseDTO{ // 返回标准的 JSON 404
				Status:  http.StatusNotFound,
				Message: "API route not found",
				Data:    gin.H{"path": path},
			})
			return
		}
		ctx.File("resource/static/install.html")
	})

	// 创建HTTP服务器实例
	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", "0.0.0.0", "8080"), // 设置监听地址和端口
		Handler: router,                                  // 设置请求处理器
	}
	return &InstallServer{
		HttpServer: httpServer,
	}
}

func (i *InstallServer) Run() {
	// 在goroutine中启动HTTP服务器
	go func() {
		// 启动服务器监听
		if err := i.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// 输出错误信息到控制台
			fmt.Printf("http server start error:%s\n", err.Error())
			// 退出程序
			os.Exit(1)
		}
	}()
	fmt.Printf("Dash install server run at %s\n", i.HttpServer.Addr)

	ok := <-quit
	if ok == "finish" {
		fmt.Printf("shutdown server ...\n")
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := i.HttpServer.Shutdown(ctx); err != nil {
			fmt.Printf("server shutdown err: %s\n", err.Error())
		}
		fmt.Printf("server exiting\n")
	}

}

func (i *InstallServer) Shutdown() {
	quit <- "finish"
}

func validateMySQLConnection(database param.Database) error {
	// 构建 MySQL DSN 连接字符串
	var dsn string
	if database.Username != nil && database.Password != nil && database.Host != nil && database.Port != nil && database.Database != nil {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			*database.Username, *database.Password, *database.Host, *database.Port, *database.Database)
	} else {
		return fmt.Errorf("MySQL 连接参数不完整")
	}
	// 尝试连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	// 获取底层的 sql.DB 实例
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %v", err)
	}
	defer sqlDB.Close()

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %v", err)
	}

	return nil
}

func validateRedisConnection(redisConfig param.Redis) error {
	// 构建 Redis 客户端配置
	opts := &redis.Options{
		Addr: fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		DB:   redisConfig.Database,
	}

	// 设置密码（如果提供）
	if redisConfig.Password != nil && *redisConfig.Password != "" {
		opts.Password = *redisConfig.Password
	}

	// 设置用户名（如果提供，Redis 6.0+ 支持）
	if redisConfig.Username != nil && *redisConfig.Username != "" {
		opts.Username = *redisConfig.Username
	}

	// 创建 Redis 客户端
	client := redis.NewClient(opts)
	defer client.Close()

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis 连接测试失败: %v", err)
	}

	return nil
}

// writeInstallYAML 将安装参数写入 conf/install.yaml 文件，并设置安装状态为 installing（并保证 status 为第一行）
// 参数：installParams - 安装参数结构体
// 返回：error - 错误信息，nil 表示成功
func writeInstallYAML(installParams *param.Install) error {
	// 获取当前工作目录
	workDir, err := os.Getwd() // 获取进程工作目录
	if err != nil {
		return fmt.Errorf("获取工作目录失败: %v", err)
	}

	// 构建 conf/install.yaml 文件路径
	installFilePath := filepath.Join(workDir, "conf", "install.yaml") // 写入 conf 目录，文件名为 install.yaml

	// 使用结构体并将 Status 放在字段首位，确保 YAML 序列化时 status 在第一行
	installConfig := struct {
		Status         string           `yaml:"status"` // 安装状态，需位于首行
		*param.Install `yaml:",inline"` // 其余安装参数内联到顶层
	}{
		Status:  "installing",  // 写入后设置状态为 installing（安装中）
		Install: installParams, // 填充安装参数
	}

	// 将参数序列化为 YAML
	yamlData, err := yaml.Marshal(installConfig) // 序列化为 YAML
	if err != nil {
		return fmt.Errorf("序列化安装参数失败: %v", err)
	}

	// 写入文件（0644 权限）
	if err := os.WriteFile(installFilePath, yamlData, 0644); err != nil { // 将 YAML 内容写入文件
		return fmt.Errorf("写入 install.yaml 文件失败: %v", err)
	}

	return nil // 成功
}

type ConfigYAML struct {
	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"server"`

	Logging struct {
		Filename string `yaml:"filename"`
		Level    struct {
			App  string `yaml:"app"`
			Gorm string `yaml:"gorm"`
		} `yaml:"level"`
		MaxSize  int  `yaml:"maxsize"`
		MaxAge   int  `yaml:"maxage"`
		Compress bool `yaml:"compress"`
	} `yaml:"logging"`

	SQLite3 struct {
		Enable   bool   `yaml:"enable"`
		Filename string `yaml:"filename"`
	} `yaml:"sqlite3"`

	MySQL struct {
		DSN string `yaml:"dsn"`
	} `yaml:"mysql"`

	Cache struct {
		Redis struct {
			Addr     string `yaml:"addr"`
			Password string `yaml:"password"`
			DB       int    `yaml:"db"`
		} `yaml:"redis"`
		DefaultTTL string `yaml:"default_ttl"`
	} `yaml:"cache"`

	Dash struct {
		LogMode string `yaml:"log_mode"`
		Mode    string `yaml:"mode"`
		WorkDir string `yaml:"work_dir"`
		LogDir  string `yaml:"log_dir"`
	} `yaml:"dash"`
}

func updateConfigYAML(installParams *param.Install) error {
	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取工作目录失败: %v", err)
	}
	// 构建配置文件路径
	configPath := filepath.Join(workDir, "conf", "config.yaml")

	// 读取现有配置文件
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析现有配置
	var config ConfigYAML
	if err = yaml.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 根据数据库类型更新配置
	dbType := strings.ToLower(installParams.Database.Type)
	switch dbType {
	case "mysql":
		// 禁用 SQLite，启用 MySQL
		config.SQLite3.Enable = false
		if installParams.Database.Username != nil && installParams.Database.Password != nil &&
			installParams.Database.Host != nil && installParams.Database.Port != nil &&
			installParams.Database.Database != nil {
			config.MySQL.DSN = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&interpolateParams=true",
				*installParams.Database.Username, *installParams.Database.Password,
				*installParams.Database.Host, *installParams.Database.Port, *installParams.Database.Database)
		}
	case "sqlite", "sqlite3":
		// 启用 SQLite，禁用 MySQL
		config.SQLite3.Enable = true
		config.MySQL.DSN = ""
	default:
		return fmt.Errorf("不支持的数据库类型: %s", installParams.Database.Type)
	}

	// 更新 Redis 配置
	config.Cache.Redis.Addr = fmt.Sprintf("%s:%d", installParams.Redis.Host, installParams.Redis.Port)
	config.Cache.Redis.DB = installParams.Redis.Database
	if installParams.Redis.Password != nil {
		config.Cache.Redis.Password = *installParams.Redis.Password
	} else {
		config.Cache.Redis.Password = ""
	}

	// 确保服务器配置为容器友好
	config.Server.Host = "0.0.0.0"
	config.Server.Port = "8080"
	config.Dash.LogMode = "console" // 便于容器日志查看

	// 序列化更新后的配置
	updatedConfigData, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 写入配置文件
	if err := os.WriteFile(configPath, updatedConfigData, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}
