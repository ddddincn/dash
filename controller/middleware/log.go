package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type GinLoggerMiddleware struct {
	logger *zap.Logger
}

func NewGinLoggerMiddleware(logger *zap.Logger) *GinLoggerMiddleware {
	return &GinLoggerMiddleware{logger: logger}
}

type GinLoggerConfig struct {
	SkipPaths []string
}

func (g *GinLoggerMiddleware) LoggerWithConfig(conf GinLoggerConfig) gin.HandlerFunc {

	logger := g.logger.WithOptions(zap.WithCaller(false))
	notLogged := conf.SkipPaths

	var skip map[string]struct{}

	if length := len(notLogged); length > 0 {
		skip = make(map[string]struct{}, length)
		for _, path := range notLogged {
			skip[path] = struct{}{}
		}
	}

	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		raw := ctx.Request.URL.RawQuery

		ctx.Next()

		if len(ctx.Errors) > 0 {
			logger.Error(ctx.Errors.ByType(gin.ErrorTypePrivate).String())
		}
		if _, ok := skip[path]; !ok {
			if raw != "" {
				path = path + "?" + raw
			}
			path = strings.ReplaceAll(path, "\n", "")
			path = strings.ReplaceAll(path, "\r", "")
			clientIP := strings.ReplaceAll(ctx.ClientIP(), "\n", "")
			clientIP = strings.ReplaceAll(clientIP, "\r", "")

			logger.Info("[GIN]",
				zap.Time("begin_time", start),
				zap.Int("status", ctx.Writer.Status()),
				zap.Duration("latency", time.Since(start)),
				zap.String("client_ip", clientIP),
				zap.String("method", ctx.Request.Method),
				zap.String("path", path),
			)
		}
	}
}

func (g *GinLoggerMiddleware) Logger() gin.HandlerFunc {
	logger := g.logger.WithOptions(zap.WithCaller(false))

	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		raw := ctx.Request.URL.RawQuery

		ctx.Next()

		if len(ctx.Errors) > 0 {
			logger.Error(ctx.Errors.ByType(gin.ErrorTypePrivate).String())
		}

		if raw != "" {
			path = path + "?" + raw
		}
		path = strings.ReplaceAll(path, "\n", "")
		path = strings.ReplaceAll(path, "\r", "")
		clientIP := strings.ReplaceAll(ctx.ClientIP(), "\n", "")
		clientIP = strings.ReplaceAll(clientIP, "\r", "")

		logger.Info("[GIN]",
			zap.Time("begin_time", start),
			zap.Int("status", ctx.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", clientIP),
			zap.String("method", ctx.Request.Method),
			zap.String("path", path),
		)

	}
}
