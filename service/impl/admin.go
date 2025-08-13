package impl

import (
	"context"
	"dash/model/entity"
	"dash/model/param"
	"dash/service"
	"dash/utils/xerr"
)

type adminServiceImpl struct {
	UserService service.UserService
}

func NewAdminService(userService service.UserService) service.AdminService {
	return &adminServiceImpl{
		UserService: userService,
	}
}

func (a *adminServiceImpl) Auth(ctx context.Context, loginParam *param.LoginParam) (*entity.User, error) {
	missMatchTip := "用户名或密码不正确"

	user, err := a.UserService.GetUserByUsername(ctx, loginParam.Username)

	// 2. 检查用户是否存在
	if xerr.GetType(err) == xerr.NoRecord {
		return nil, xerr.WithMsg(err, missMatchTip).WithStatus(xerr.StatusBadRequest)
	}

	// 3. 检查用户是否过期
	err = a.UserService.MustNotExpire(ctx, user.ExpireTime)
	if err != nil {
		return nil, err
	}

	// 4. 验证密码（使用bcrypt）
	if !a.UserService.PasswordMatch(ctx, user.Password, loginParam.Password) {
		return nil, xerr.BadParam.New("").WithMsg(missMatchTip).WithStatus(xerr.StatusBadRequest)
	}

	return user, nil
}

// func (a *adminServiceImpl) buildAuthToken(user *entity.User) *dto.AuthToken {
// 	// 1. 生成UUID格式的Token
// 	accessToken := uuid.New().String()  // 访问令牌
// 	refreshToken := uuid.New().String() // 刷新令牌

// 	// 2. 构建返回对象
// 	authToken := &dto.AuthToken{
// 		AccessToken:  accessToken,
// 		ExpiredIn:    consts.AccessTokenExpiredSeconds, // 24小时
// 		RefreshToken: refreshToken,
// 	}

// 	// 3. 双向缓存存储（Token->UserID 和 UserID->Token）
// 	// Token到用户ID的映射
// 	cache.Set(cache.BuildTokenAccessKey(accessToken), user.ID,
// 		time.Second*consts.AccessTokenExpiredSeconds)
// 	cache.Set(cache.BuildTokenRefreshKey(refreshToken), user.ID,
// 		consts.RefreshTokenExpiredDays*24*3600*time.Second)

// 	// 用户ID到Token的映射
// 	cache.Set(cache.BuildAccessTokenKey(user.ID), accessToken,
// 		time.Second*consts.AccessTokenExpiredSeconds)
// 	cache.Set(cache.BuildRefreshTokenKey(user.ID), refreshToken,
// 		consts.RefreshTokenExpiredDays*24*3600*time.Second)

// 	return authToken
// }
