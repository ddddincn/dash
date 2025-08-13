package service

import (
	"context"
	"dash/model/entity"
	"dash/model/param"
)

type AdminService interface {
	Auth(ctx context.Context, loginParam *param.LoginParam) (*entity.User, error)
}
