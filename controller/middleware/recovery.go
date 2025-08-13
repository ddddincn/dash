package middleware

import (
	"errors"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"dash/model/dto"
)

type RecoveryMiddleware struct {
	logger *zap.Logger
}

func NewRecoveryMiddleware(logger *zap.Logger) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		logger: logger,
	}
}

func (r *RecoveryMiddleware) RecoveryWithLogger() gin.HandlerFunc {
	logger := r.logger.WithOptions(zap.AddCallerSkip(2))

	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne, &se) {
						seStr := strings.ToLower(se.Error())
						if strings.Contains(seStr, "broken pipe") ||
							strings.Contains(seStr, "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				if brokenPipe {
					logger.Error(ctx.Request.URL.Path, zap.Any("error", err))
				} else {
					logger.DPanic("[Recovery]  panic recovered", zap.Any("error", err))
				}

				if brokenPipe {
					ctx.Error(err.(error))
					ctx.Abort()
				} else {
					code := http.StatusInternalServerError
					ctx.AbortWithStatusJSON(code, &dto.BaseDTO{Status: code, Message: http.StatusText(code)})
				}
			}
		}()
		ctx.Next()
	}
}
