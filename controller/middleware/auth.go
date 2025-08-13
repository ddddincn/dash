package middleware

import (
	"dash/cache"
	"dash/consts"
	"dash/model/dto"
	"dash/model/property"
	"dash/service"
	"dash/utils/xerr"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	OptionService       service.OptionService
	OneTimeTokenService service.OneTimeTokenService
	UserService         service.UserService
}

func NewAuthMiddleware(optionService service.OptionService, oneTimeTokenService service.OneTimeTokenService, userService service.UserService) *AuthMiddleware {
	authMiddleware := &AuthMiddleware{
		OptionService:       optionService,
		OneTimeTokenService: oneTimeTokenService,
		UserService:         userService,
	}
	return authMiddleware
}

func (a *AuthMiddleware) GetWrapHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		isInstalled, err := a.OptionService.GetOrByDefaultWithErr(ctx, property.IsInstalled, false)
		if err != nil {
			abortWithStatusJSON(ctx, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
		if !isInstalled.(bool) {
			abortWithStatusJSON(ctx, http.StatusBadRequest, "Blog is not initialized")
			return
		}

		oneTimeToken, ok := ctx.GetQuery(consts.OneTimeTokenQueryName)
		if ok {
			allowedURL, ok := a.OneTimeTokenService.Get(oneTimeToken)
			if !ok {
				abortWithStatusJSON(ctx, http.StatusBadRequest, "OneTimeToken is not exist or expired")
				return
			}
			currentURL := ctx.Request.URL.Path
			if currentURL != allowedURL {
				abortWithStatusJSON(ctx, http.StatusBadRequest, "The one-time token does not correspond the request uri")
				return
			}
			return
		}

		tokenWithBearer := ctx.GetHeader(consts.AdminTokenHeaderName)
		if tokenWithBearer == "" {
			abortWithStatusJSON(ctx, http.StatusUnauthorized, "未登录，请登录后访问")
			return
		}
		token := tokenWithBearer[7:]
		userID, err := cache.Redis.Get(ctx, cache.BuildTokenAccessKey(token)).Result()

		if err != nil {
			abortWithStatusJSON(ctx, http.StatusUnauthorized, "Token 已过期或不存在")
			return
		}
		userIDInt, err := strconv.Atoi(userID)
		if err != nil {
			abortWithStatusJSON(ctx, http.StatusUnauthorized, "Token 已过期或不存在")
			return
		}
		user, err := a.UserService.GetUserByID(ctx, int32(userIDInt))
		if xerr.GetType(err) == xerr.NoRecord {
			_ = ctx.Error(err)
			abortWithStatusJSON(ctx, http.StatusUnauthorized, "用户不存在")
			return
		}
		if err != nil {
			_ = ctx.Error(err)
			abortWithStatusJSON(ctx, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
		ctx.Set(consts.AuthorizedUser, user)
	}
}

func abortWithStatusJSON(ctx *gin.Context, status int, message string) {
	ctx.AbortWithStatusJSON(200, &dto.BaseDTO{
		Status:  status,
		Message: message,
	})
}
