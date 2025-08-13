package controller

import (
	"dash/config"
	"dash/controller/middleware"
	"dash/model/dto"
	"net/http"
	"strings"

	// 引入 strings 用于判断路径前缀

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// registerRouter 注册路由和中间件
// 配置CORS、日志、恢复中间件以及业务路由
func (s *Server) registerRouter() {
	router := s.Router
	// 开发模式下配置CORS中间件解决跨域问题
	if config.IsDev() {
		cfg := cors.Config{ // 创建 CORS 配置
			AllowMethods: []string{"PUT", "PATCH", "GET", "DELETE", "POST", "OPTIONS"}, // 允许的方法
			AllowHeaders: []string{
				"Origin",
				"Authorization",
				"Content-Type",
				"Accept",
				"X-Requested-With",
				"Access-Control-Request-Method",
				"Access-Control-Request-Headers",
			},
			AllowCredentials: true, // 允许携带凭证（如 Cookie）
			ExposeHeaders: []string{
				"Content-Length",
				"Content-Type",
				"Content-Disposition",
				"Access-Control-Allow-Origin",
			},
		}
		// 允许任意 localhost 和 127.0.0.1 端口来源，便于前端在不同端口启动（如 5137/5173 等）
		cfg.AllowOriginFunc = func(origin string) bool { // 自定义来源校验函数
			return strings.HasPrefix(origin, "http://localhost:") || // 放行 localhost 任意端口
				strings.HasPrefix(origin, "http://127.0.0.1:") // 放行 127.0.0.1 任意端口
		}
		router.Use(cors.New(cfg)) // 注册 CORS 中间件
	}

	// 创建并注册中间件
	ginLoggerMiddleware := middleware.NewGinLoggerMiddleware(s.Logger)                // 创建日志中间件
	recoveryMiddleware := middleware.NewRecoveryMiddleware(s.Logger)                  // 创建恢复中间件
	router.Use(ginLoggerMiddleware.Logger(), recoveryMiddleware.RecoveryWithLogger()) // 注册中间件
	// 健康检查路由
	router.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, dto.BaseDTO{
			Status:  http.StatusOK,
			Message: "OK",
			Data:    gin.H{"message": "pong"},
		})
	})
	staticRouter := router.Group("/")
	{
		staticRouter.StaticFile("", "resource/static/index.html")
		staticRouter.StaticFS("/assets", gin.Dir("resource/static/assets", false)) // 挂载静态资源目录（JS/CSS/图片等）
	}
	publicRouter := router.Group("/api")
	{

		publicThemeRouter := publicRouter.Group("/theme")
		{
			publicThemeRouter.GET("/:themeID", s.handler(s.ThemeHandler.GetThemeSettings))
		}
		publicPostRouter := publicRouter.Group("/posts")
		{
			publicPostRouter.GET("", s.handler(s.PostHandler.ListPosts))
			publicPostRouter.GET("/:slug", s.handler(s.PostHandler.GetPostBySlug))
			publicPostRouter.GET("/search", s.handler(s.PostHandler.SearchPost))
			publicPostRouter.GET("/archive", s.handler(s.PostHandler.GetPostArchive))
		}
		publicMenuRouter := publicRouter.Group("/menus")
		{
			publicMenuRouter.GET("", s.handler(s.MenuHandler.ListMenus))
		}
		publicCategoryRouter := publicRouter.Group("/categories")
		{
			publicCategoryRouter.GET("", s.handler(s.CategoryHandler.ListCategoriesWithPosts))
			publicCategoryRouter.GET("/:slug/posts", s.handler(s.CategoryHandler.ListPostsByCategorySlug))
		}
		publicTagRouter := publicRouter.Group("/tags")
		{
			publicTagRouter.GET("", s.handler(s.TagHandler.ListTagsWithPosts))
			publicTagRouter.GET("/:slug/posts", s.handler(s.TagHandler.ListPostsByTagSlug))
			publicTagRouter.GET("/count", s.handler(s.TagHandler.ListTags))

		}
		publicSheetRouter := publicRouter.Group("/sheet")
		{
			publicSheetRouter.GET("/:slug", s.handler(s.PostHandler.GetPostBySlug))
		}

	}
	adminRouter := router.Group("/api/admin")
	{

		// adminRouter.GET("/is_install", s.handler(s.InstallHandler.IsInstall))
		// adminRouter.POST("/install", s.handler(s.InstallHandler.InstallBlog))
		adminAuthRouter := adminRouter.Group("/auth")
		{
			adminAuthRouter.POST("/login", s.handler(s.AdminHandler.Login))
			adminAuthRouter.POST("/refresh", s.handler(s.AdminHandler.Refresh))
		}
		adminStatisticRouter := adminRouter.Group("/statistics").Use(s.AuthMiddleware.GetWrapHandler())
		{
			adminStatisticRouter.GET("", s.handler(s.StatisticHandler.Statistic))
		}
		adminPostsRouter := adminRouter.Group("/posts").Use(s.AuthMiddleware.GetWrapHandler())
		{
			adminPostsRouter.GET("", s.handler(s.PostHandler.ListPosts))
			adminPostsRouter.GET("/:id", s.handler(s.PostHandler.GetPostByID))
			adminPostsRouter.GET("/slug/:slug", s.handler(s.PostHandler.GetPostBySlug))
			adminPostsRouter.POST("", s.handler(s.PostHandler.CreatePost))
			adminPostsRouter.PUT("/:id", s.handler(s.PostHandler.UpdatePost))
			adminPostsRouter.PATCH("/:id/status/:status", s.handler(s.PostHandler.UpdatePostStatus))
			adminPostsRouter.PATCH("/status/:status", s.handler(s.PostHandler.UpdatePostStatusBatch))
			adminPostsRouter.DELETE("/:id", s.handler((s.PostHandler.DeletePost)))
			adminPostsRouter.DELETE("", s.handler((s.PostHandler.DeletePostBatch)))
		}
		adminCategoryRouter := adminRouter.Group("/categories").Use(s.AuthMiddleware.GetWrapHandler())
		{
			adminCategoryRouter.GET("", s.handler(s.CategoryHandler.ListCategories))
			adminCategoryRouter.GET("/:id", s.handler(s.CategoryHandler.GetCategoryByID))
			adminCategoryRouter.POST("", s.handler(s.CategoryHandler.CreateCategory))
			adminCategoryRouter.PUT("/:id", s.handler(s.CategoryHandler.UpdateCategory))
			adminCategoryRouter.DELETE("/:id", s.handler(s.CategoryHandler.DeleteCategory))
		}
		adminTagRouter := adminRouter.Group("/tags").Use(s.AuthMiddleware.GetWrapHandler())
		{
			adminTagRouter.GET("", s.handler(s.TagHandler.ListTags))
			adminTagRouter.GET("/:id", s.handler(s.TagHandler.GetTagByID))
			adminTagRouter.POST("", s.handler(s.TagHandler.CreateTag))
			adminTagRouter.PUT("/:id", s.handler(s.TagHandler.UpdateTag))
			adminTagRouter.DELETE("/:id", s.handler(s.TagHandler.DeleteTag))
		}
	}

	// NoRoute 回退：
	// - 对于以 /api 开头的未知接口，返回 404 JSON，便于前端识别接口不存在
	// - 对于其他任意未知路径（如 /console、/about 等），统一回退到打包后的 index.html，交给前端路由处理
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
		// 非 /api 的路径统一回退至 index.html（React SPA 入口）
		ctx.File("resource/static/index.html")
	})
}
