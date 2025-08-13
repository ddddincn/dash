package service

import (
	"context"
	"dash/consts"
	"dash/model/entity"
)

type PostCategoryService interface {
	ListByPostIDs(ctx context.Context, postIDs []int32) ([]*entity.PostCategory, error)
	ListCategoriesByPostID(ctx context.Context, postID int32) ([]*entity.Category, error)
	ListPostsByCategoryID(ctx context.Context, categoryID int32, status consts.PostStatus) ([]*entity.Post, error)
	ListCategoryMapByPostID(ctx context.Context, postIDs []int32) (map[int32][]*entity.Category, error)
}
