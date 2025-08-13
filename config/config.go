package config

import (
	"dash/utils"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// NewConfig 读取配置文件并完成必要的初始化工作
func NewConfig() *Config {
	var configFile string
	// 提供从参数-config=config_file_path
	flag.StringVar(&configFile, "config", "", "")
	flag.Parse()

	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetConfigType("yaml")
	if configFile != "" { // 如果提供了配置文件参数就使用参数路径
		viper.SetConfigFile(configFile)
	} else { // 否则使用默认配置文件路径
		viper.AddConfigPath("./conf")
		viper.SetConfigName("config")
	}
	// 设置管理员路由
	viper.SetDefault("dash.admin_url_path", "admin")
	// 读取配置文件并解析到conf结构体中
	conf := &Config{}
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := viper.Unmarshal(conf); err != nil {
		panic(err)
	}
	// 检查必要参数
	if conf.Dash.WorkDir == "" {
		//如果workDir为空就设置为当前目录
		pwd, err := os.Getwd()
		if err != nil {
			panic(errors.Wrap(err, "init config get current dir"))
		}
		conf.Dash.WorkDir, _ = filepath.Abs(pwd)
	} else {
		// 如果workDir不为空就取参数目录的绝对路径
		workDir, err := filepath.Abs(conf.Dash.WorkDir)
		if err != nil {
			panic(err)
		}
		conf.Dash.WorkDir = workDir
	}
	//拼接路径函数
	normalizeDir := func(path *string, subDir string) {
		if *path == "" {
			*path = filepath.Join(conf.Dash.WorkDir, subDir)
		} else {
			temp, err := filepath.Abs(*path)
			if err != nil {
				panic(err)
			}
			*path = temp
		}
	}

	normalizeDir(&conf.Dash.LogDir, "log")
	// normalizeDir(&conf.Dash.UploadDir, consts.DashUploadDir)
	// 查看sqlite是否启用，如果启用还需要创建sqliteDB
	if conf.SQLite3 != nil && conf.SQLite3.Enable {
		normalizeDir(&conf.SQLite3.File, "dash.db")
	}
	// 初始化目录
	initDirectory(conf)
	mode = conf.Dash.Mode
	logMode = conf.Dash.LogMode
	return conf
}

func initDirectory(conf *Config) {
	if err := utils.MakeDir(conf.Dash.LogDir); err != nil {
		panic(fmt.Errorf("initDirectory err=%w", err))
	}
}

var (
	mode    string
	logMode LogMode
)

func IsDev() bool {
	return mode == "development"
}

func LogToConsole() bool {
	switch logMode {
	case Console:
		return true
	case File:
		return false
	default:
		return IsDev()
	}
}
