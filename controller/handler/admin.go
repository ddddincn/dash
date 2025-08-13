package handler

import (
	"dash/consts"
	"dash/model/dto"
	"dash/model/param"
	"dash/service"
	"dash/utils/xerr"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AdminHandler struct {
	AdminService service.AdminService
	JWTService   service.JWTService
}

func NewAdminHandler(adminService service.AdminService, jwtService service.JWTService) *AdminHandler {
	return &AdminHandler{
		AdminService: adminService,
		JWTService:   jwtService,
	}
}

func (a *AdminHandler) Login(ctx *gin.Context) (interface{}, error) {
	loginParam := &param.LoginParam{}
	err := ctx.ShouldBindJSON(loginParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(e.Error())
		}
		return nil, xerr.BadParam.Wrapf(err, "invalid parameter").WithStatus(xerr.StatusBadRequest).WithMsg("invalid parameter")
	}
	user, err := a.AdminService.Auth(ctx, loginParam)
	if err != nil {
		return nil, err
	}
	accessToken, refreshToken, err := a.JWTService.GenerateTokens(user)
	if err != nil {
		return nil, err
	}
	token := &dto.AccessToken{
		AccessToken: accessToken,
		ExpiredIn:   int(time.Now().Add(consts.AccessTokenExpiredSeconds * time.Second).UnixMilli()),
	}
	ctx.SetCookie("refresh_token", refreshToken, int(consts.RefreshTokenExpiredDays)*24*3600, "/api/admin/auth", "", true, true)
	return token, nil
}

func (a *AdminHandler) Refresh(ctx *gin.Context) (interface{}, error) {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		return nil, xerr.Forbidden.New("no refresh_token").WithStatus(xerr.StatusUnauthorized).WithMsg("登录已过期，需重新登录")
	}
	accessToken, err := a.JWTService.RefreshToken(refreshToken)
	if err != nil {
		return nil, xerr.BadParam.Wrap(err).WithStatus(xerr.StatusUnauthorized).WithMsg("登录已过期，需重新登录")
	}
	token := &dto.AccessToken{
		AccessToken: accessToken,
		ExpiredIn:   int(time.Now().Add(consts.AccessTokenExpiredSeconds * time.Second).UnixMilli()),
	}
	return token, nil
}
