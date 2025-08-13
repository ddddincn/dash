package service

import (
	"context"
	"dash/consts"
	"dash/model/dto"
	"dash/model/entity"
	"dash/model/param"
)

type PostTagService interface {
	ListTagsByPostID(ctx context.Context, postID int32) ([]*entity.Tag, error)
	ListPostsByTagID(ctx context.Context, tagID int32, status consts.PostStatus) ([]*entity.Post, error)
	ListTagMapByPostID(ctx context.Context, postIDs []int32) (map[int32][]*entity.Tag, error)
	ListTagWithPostCount(ctx context.Context, sort *param.Sort) ([]*dto.TagWithPostCount, error)
}
