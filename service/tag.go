package service

import (
	"context"
	"dash/model/dto"
	"dash/model/entity"
	"dash/model/param"
)

type TagService interface {
	Create(ctx context.Context, tagParam *param.Tag) (*entity.Tag, error)
	DeleteByID(ctx context.Context, id int32) error
	UpdateByID(ctx context.Context, id int32, tagParam *param.Tag) (*entity.Tag, error)
	List(ctx context.Context, sort *param.Sort) ([]*entity.Tag, error)
	ListByIDs(ctx context.Context, tagIDs []int32) ([]*entity.Tag, error)
	GetTagByID(ctx context.Context, id int32) (*entity.Tag, error)
	GetTagByName(ctx context.Context, name string) (*entity.Tag, error)
	GetTagBySlug(ctx context.Context, slug string) (*entity.Tag, error)
	GetTagsCount(ctx context.Context) (int64, error)

	ConvertToTagDTO(ctx context.Context, tag *entity.Tag) (*dto.Tag, error)
	ConvertToTagDTOs(ctx context.Context, tags []*entity.Tag) ([]*dto.Tag, error)
	ConvertToTagWithPostCountDTO(ctx context.Context, tag *entity.Tag) (*dto.TagWithPostCount, error)
	ConvertToTagWithPostCountDTOs(ctx context.Context, tags []*entity.Tag) ([]*dto.TagWithPostCount, error)
}
