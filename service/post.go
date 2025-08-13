package service

import (
	"context" // 常量定义
	"dash/consts"
	"dash/model/entity" // 实体模型
	"dash/model/param"  // 参数模型
)

type PostService interface {
	BasePostService
	Page(ctx context.Context, postQuery param.PostQuery) ([]*entity.Post, int64, error)
	GetPrevPosts(ctx context.Context, post *entity.Post, size int) ([]*entity.Post, error)
	GetNextPosts(ctx context.Context, post *entity.Post, size int) ([]*entity.Post, error)
	GetPostCountByStatus(ctx context.Context, status consts.PostStatus) (int64, error)
	GetVisitCount(ctx context.Context) (int64, error)
	GetLikeCount(ctx context.Context) (int64, error)
}
