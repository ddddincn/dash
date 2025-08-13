package main

import (
	"dash/controller"
	"dash/injection"
	"fmt"

	"github.com/spf13/viper"
)

type App struct {
	Server interface{}
}

func NewApp() *App {
	v := viper.New()
	v.SetConfigName("install")
	v.SetConfigType("yaml")
	v.AddConfigPath("conf")
	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("read install config file failed: %w", err))
	}
	status := v.GetString("status")
	if status == "installed" { // 如果已经安装就返回dashServer实例
		return &App{
			Server: injection.NewDashServer(),
		}
	}
	return &App{ // 如果没有安装就返回installServer实例
		Server: controller.NewInstallServer(),
	}
}

func (a *App) Run() {
	switch a.Server.(type) {
	case *controller.InstallServer: // 如果是installServer实例就运行installServer
		a.Server.(*controller.InstallServer).Run() //安装完成后会自动退出安装server
		dashServer := injection.NewDashServer()    //然后启动dashServer，进入dashServer初始化
		dashServer.Install()
		dashServer.Run()
	case *controller.Server: // 如果是dashServer实例就运行dashServer
		a.Server.(*controller.Server).Run()
	default:
		panic("app server type not support")
	}
}

func main() {
	app := NewApp()
	app.Run()
}
