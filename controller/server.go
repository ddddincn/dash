package controller

import (
	"context"
	"dash/cache"
	"dash/config"
	"dash/controller/handler"
	"dash/controller/middleware"
	"dash/model/dto"
	"dash/model/param"
	"dash/utils/xerr"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type Server struct {
	Conf       *config.Config // 应用配置信息
	Logger     *zap.Logger    // 日志记录器
	DB         *gorm.DB
	Cache      *cache.RedisCache
	Router     *gin.Engine  // Gin路由引擎
	HttpServer *http.Server // HTTP服务器实例

	AuthMiddleware *middleware.AuthMiddleware

	PostHandler      *handler.PostHandler
	CategoryHandler  *handler.CategoryHandler // 分类处理器
	TagHandler       *handler.TagHandler
	StatisticHandler *handler.StatisticsHandler
	ThemeHandler     *handler.ThemeHandler
	MenuHandler      *handler.MenuHandler
	AdminHandler     *handler.AdminHandler
	InstallHandler   *handler.InstallHandler
}

func NewServer(
	conf *config.Config,
	logger *zap.Logger,
	db *gorm.DB,
	cache *cache.RedisCache,
	authMiddleware *middleware.AuthMiddleware,

	postHandler *handler.PostHandler,
	categoryHandler *handler.CategoryHandler,
	tagHandler *handler.TagHandler,
	statisticHandler *handler.StatisticsHandler,
	themeHandler *handler.ThemeHandler,
	menuHandler *handler.MenuHandler,
	adminHandler *handler.AdminHandler,
	installHandler *handler.InstallHandler,
) *Server {
	// 根据环境设置Gin模式
	if !config.IsDev() {
		gin.SetMode(gin.ReleaseMode)
	}
	// 创建Gin路由引擎
	router := gin.New()

	// 创建HTTP服务器实例
	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", conf.Server.Host, conf.Server.Port), // 设置监听地址和端口
		Handler: router,                                                   // 设置请求处理器
	}

	// 创建服务器实例
	server := &Server{
		Conf:       conf,   // 设置配置信息
		Logger:     logger, // 设置日志记录器
		DB:         db,
		Cache:      cache,
		Router:     router,     // 设置路由引擎
		HttpServer: httpServer, // 设置HTTP服务器

		AuthMiddleware: authMiddleware,

		PostHandler:      postHandler,
		CategoryHandler:  categoryHandler, // 设置分类处理器
		TagHandler:       tagHandler,
		StatisticHandler: statisticHandler,
		ThemeHandler:     themeHandler,
		MenuHandler:      menuHandler,
		AdminHandler:     adminHandler,
		InstallHandler:   installHandler,
	}

	// 注册路由
	server.registerRouter()
	return server
}

func (s *Server) Install() {
	v := viper.New()
	v.SetConfigName("install")
	v.SetConfigType("yaml")
	v.AddConfigPath(path.Join(s.Conf.Dash.WorkDir, "conf"))
	if err := v.ReadInConfig(); err != nil {
		s.Logger.Fatal("read install config file failed", zap.Error(err))
	}

	installParams := &param.Install{
		User: param.User{
			Username: v.GetString("user.username"), // 用户名
			Password: v.GetString("user.password"), // 密码
			Nickname: v.GetString("user.nickname"), // 昵称
			Email:    v.GetString("user.email"),    // 邮箱
		},
		Title: v.GetString("title"), // 站点标题
		URL:   v.GetString("url"),   // 站点 URL
	}
	_, err := s.InstallHandler.InstallBlog(installParams)
	if err != nil {
		s.Logger.Fatal("install blog failed", zap.Error(err))
	}

	// 安装成功后将 conf/install.yaml 中的 status 更新为 installed
	installFilePath := path.Join(s.Conf.Dash.WorkDir, "conf", "install.yaml") // 计算文件路径
	if err := setInstallStatusInstalled(installFilePath); err != nil {        // 调用工具函数更新状态
		s.Logger.Error("update install status failed", zap.Error(err)) // 仅记录错误，不中断
	}
}

func setInstallStatusInstalled(filePath string) error {
	data, err := os.ReadFile(filePath) // 读取文件
	if err != nil {
		return fmt.Errorf("读取 install.yaml 失败: %v", err)
	}

	vv := viper.New()
	vv.SetConfigType("yaml")
	if err := vv.ReadConfig(strings.NewReader(string(data))); err != nil {
		return fmt.Errorf("解析 install.yaml 失败: %v", err)
	}

	// 将配置反序列化为 map，便于重组并控制字段顺序
	m := map[string]any{}                            // 用于接收顶层键值
	if err := yaml.Unmarshal(data, &m); err != nil { // 反序列化到 map
		return fmt.Errorf("反序列化 install.yaml 失败: %v", err)
	}

	// 更新状态
	m["status"] = "installed" // 将状态置为 installed

	// 重新组装 YAML，强制 status 放在第一行
	// 方案：先写出仅包含 status 的 YAML，再写出其他键，保持顶层扁平结构
	var b strings.Builder                // 字符串构建器
	b.WriteString("status: installed\n") // 第一行写入状态
	for k, v := range m {                // 遍历其余键
		if k == "status" { // 跳过 status 已写
			continue
		}
		out, err := yaml.Marshal(map[string]any{k: v}) // 单独序列化每一个顶层键
		if err != nil {
			return fmt.Errorf("序列化键 %s 失败: %v", k, err)
		}
		b.Write(out) // 追加序列化结果
	}

	// 写回文件
	if err := os.WriteFile(filePath, []byte(b.String()), 0644); err != nil { // 覆盖写入
		return fmt.Errorf("写入 install.yaml 失败: %v", err)
	}
	return nil // 成功
}

func (s *Server) Run() {

	// 在goroutine中启动HTTP服务器
	go func() {
		// 启动服务器监听
		if err := s.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// 记录服务器启动错误
			s.Logger.Error("unexpected error from ListenAndServe", zap.Error(err))
			// 输出错误信息到控制台
			fmt.Printf("http server start error:%s\n", err.Error())
			// 退出程序
			os.Exit(1)
		}
	}()
	s.Logger.Info(fmt.Sprintf("Dash backend server run at %s\n", s.HttpServer.Addr))
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no params) by default sends syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	s.Logger.Info("shutdown server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.HttpServer.Shutdown(ctx); err != nil {
		s.Logger.Error("server shutdown err", zap.Error(err))
	}
	s.Logger.Info("server exiting")
}

type h func(ctx *gin.Context) (interface{}, error)

func (s *Server) handler(h h) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 调用业务处理器函数
		data, err := h(ctx)
		if err != nil {
			// 记录错误日志
			s.Logger.Error("handler error", zap.Error(err))
			// 获取HTTP状态码
			status := xerr.GetHTTPStatus(err)
			// 返回错误响应
			ctx.JSON(200, &dto.BaseDTO{Status: status, Message: xerr.GetMessage(err)})
			return
		}

		// 返回成功响应
		ctx.JSON(http.StatusOK, &dto.BaseDTO{
			Status:  http.StatusOK, // 设置状态码
			Data:    data,          // 设置响应数据
			Message: "OK",          // 设置响应消息
		})
	}
}
