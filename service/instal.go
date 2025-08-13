package service

import (
	"context"
	"dash/model/param"
)

type InstallService interface {
	InstallBlog(ctx context.Context, installParam *param.Install) error
}
