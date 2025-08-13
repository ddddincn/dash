package service

import (
	"context"
	"dash/model/dto"
	"dash/model/entity"
	"dash/model/param"
	"time"
)

type UserService interface {
	Create(ctx context.Context, userParam *param.User) (*entity.User, error)
	List(ctx context.Context) ([]*entity.User, error)
	GetFirst(ctx context.Context) (*entity.User, error)
	GetUserByID(ctx context.Context, id int32) (*entity.User, error)
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)
	ConvertToUserDTO(user *entity.User) *dto.User
	ConvertToUserDTOs(users []*entity.User) []*dto.User

	EncryptPassword(ctx context.Context, plainPassword string) string
	PasswordMatch(ctx context.Context, hashedPassword, plainPassword string) bool
	MustNotExpire(ctx context.Context, expireTime *time.Time) error
}
