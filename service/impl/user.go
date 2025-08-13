package impl

import (
	"context"
	"dash/dal"
	"dash/log"
	"dash/model/dto"
	"dash/model/entity"
	"dash/model/param"
	"dash/service"
	"dash/utils"
	"dash/utils/xerr"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type userServiceImpl struct {
}

func NewUserService() service.UserService {
	return &userServiceImpl{}
}

func (u *userServiceImpl) Create(ctx context.Context, userParam *param.User) (*entity.User, error) {
	if len(userParam.Password) < 8 || len(userParam.Password) > 100 {
		return nil, xerr.BadParam.Wrap(nil).WithMsg("password length err")
	}
	user := &entity.User{
		CreateTime:  time.Now(),
		Description: userParam.Description,
		Email:       userParam.Email,
		Password:    u.EncryptPassword(ctx, userParam.Password),
		Username:    userParam.Username,
		Nickname:    userParam.Nickname,
		Avatar:      userParam.Avatar,
	}
	userDAL := dal.GetQueryByCtx(ctx).User
	err := userDAL.WithContext(ctx).Create(user)
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return user, nil
}

func (u *userServiceImpl) List(ctx context.Context) ([]*entity.User, error) {
	userDAL := dal.GetQueryByCtx(ctx).User
	users, err := userDAL.WithContext(ctx).Find()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return users, nil
}

func (u *userServiceImpl) GetFirst(ctx context.Context) (*entity.User, error) {
	userDAL := dal.GetQueryByCtx(ctx).User
	user, err := userDAL.WithContext(ctx).First()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return user, nil
}

func (u *userServiceImpl) GetUserByID(ctx context.Context, id int32) (*entity.User, error) {
	userDAL := dal.GetQueryByCtx(ctx).User
	user, err := userDAL.WithContext(ctx).Where(userDAL.ID.Eq(id)).First()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return user, nil
}

func (u *userServiceImpl) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	userDAL := dal.GetQueryByCtx(ctx).User
	user, err := userDAL.WithContext(ctx).Where(userDAL.Username.Eq(username)).First()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return user, nil
}

func (u *userServiceImpl) ConvertToUserDTO(user *entity.User) *dto.User {
	userDTO := &dto.User{
		ID:          user.ID,
		Username:    user.Username,
		Nickname:    user.Nickname,
		Email:       user.Email,
		Avatar:      user.Avatar,
		Description: user.Description,
		MFAType:     user.MfaType,
		CreateTime:  user.CreateTime.UnixMilli(),
	}
	if user.UpdateTime != nil {
		userDTO.UpdateTime = user.UpdateTime.UnixMilli()
	}
	return userDTO
}

func (u *userServiceImpl) ConvertToUserDTOs(users []*entity.User) []*dto.User {
	userDTOs := make([]*dto.User, 0, len(users))
	for _, user := range users {
		userDTOs = append(userDTOs, u.ConvertToUserDTO(user))
	}
	return userDTOs
}

func (u *userServiceImpl) EncryptPassword(ctx context.Context, plainPassword string) string {
	password, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		log.CtxError(ctx, "encrypt password", zap.Error(err))
	}
	return string(password)
}

func (u *userServiceImpl) PasswordMatch(ctx context.Context, hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

func (u *userServiceImpl) MustNotExpire(ctx context.Context, expireTime *time.Time) error {
	if expireTime == nil {
		return nil
	}
	now := time.Now()
	if expireTime.After(now) {
		return xerr.Forbidden.New("账号已被停用，请 %s 后重试", utils.TimeFormat(int(expireTime.Sub(now).Seconds()))).WithStatus(xerr.StatusForbidden)
	}
	return nil
}
