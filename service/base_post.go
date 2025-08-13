package service

import (
	"context"

	"dash/consts"
	"dash/model/entity"
	"dash/model/param"
)

type BasePostService interface {
	Create(ctx context.Context, postParam *param.Post, postType consts.PostType) (*entity.Post, error)
	DeleteByID(ctx context.Context, id int32) error
	DeleteBatchByID(ctx context.Context, ids []int32) error
	UpdateByID(ctx context.Context, id int32, postParam *param.Post, postType consts.PostType) (*entity.Post, error)
	UpdateStatusByID(ctx context.Context, id int32, status consts.PostStatus) (*entity.Post, error)
	UpdateStatusBatch(ctx context.Context, ids []int32, status consts.PostStatus) ([]*entity.Post, error)
	List(ctx context.Context, sort *param.Sort) ([]*entity.Post, error)
	ListByIDs(ctx context.Context, ids []int32) ([]*entity.Post, error)
	GetPostByID(ctx context.Context, id int32) (*entity.Post, error)
	GetPostBySlug(ctx context.Context, slug string) (*entity.Post, error)
	GetPostsCount(ctx context.Context) (int64, error)

	ConvertToEntity(ctx context.Context, postParam *param.Post, postType consts.PostType) (*entity.Post, error)
}
