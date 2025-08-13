//go:build wireinject
// +build wireinject

package injection

import (
	"dash/cache"
	"dash/config"
	"dash/controller"
	"dash/dal"

	"dash/controller/handler"
	"dash/controller/middleware"
	"dash/log"
	"dash/service/assembler"
	"dash/service/impl"

	"github.com/google/wire"
)

// InitializeApp 初始化应用程序的依赖注入
// 使用 Wire 框架自动生成依赖注入代码
// 返回值: 配置好的服务器实例
func NewDashServer() *controller.Server {
	wire.Build(
		// 基础配置和日志
		config.NewConfig,
		log.NewLogger,
		log.NewGormLogger,

		// 基础服务
		cache.NewRedisCache,
		dal.NewGormDB,

		// 业务服务
		impl.NewOptionService,
		impl.NewBasePostService,
		impl.NewPostService,
		impl.NewTagService,
		impl.NewCategoryService,
		impl.NewPostTagService,
		impl.NewPostCategoryService,
		impl.NewUserService,
		impl.NewThemeService,
		impl.NewMenuService,
		impl.NewAdminService,
		impl.NewJWTService,
		impl.NewOneTimeTokenService,
		impl.NewInstallService,

		// 组装器
		assembler.NewBasePostAssembler,
		assembler.NewPostAssembler,

		// 处理器
		handler.NewCategoryHandler,
		handler.NewTagHandler,
		handler.NewPostHandler,
		handler.NewStatisticsHandler,

		handler.NewThemeHandler,
		handler.NewMenuHandler,

		handler.NewAdminHandler,
		handler.NewInstallHandler,
		controller.NewServer,
		middleware.NewAuthMiddleware,
	)
	return nil
}
