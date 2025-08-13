package service

import (
	"context"
	"dash/model/dto"
	"dash/model/entity"
	"dash/model/param"
)

type CategoryService interface {
	Create(ctx context.Context, categoryParam *param.Category) (*entity.Category, error)
	DeleteByID(ctx context.Context, id int32) error
	UpdateByID(ctx context.Context, id int32, categoryParam *param.Category) (*entity.Category, error)
	List(ctx context.Context, sort *param.Sort) ([]*entity.Category, error)
	ListByIDs(ctx context.Context, ids []int32) ([]*entity.Category, error)
	GetCategoryByID(ctx context.Context, id int32) (*entity.Category, error)
	GetCategoryByName(ctx context.Context, name string) (*entity.Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*entity.Category, error)
	GetCategoriesCount(ctx context.Context) (int64, error)

	ConvertToCategoryDTO(ctx context.Context, category *entity.Category) (*dto.Category, error)
	ConvertToCategoryDTOs(ctx context.Context, categories []*entity.Category) ([]*dto.Category, error)
	ConvertToCategoryWithPostCountDTO(ctx context.Context, category *entity.Category) (*dto.CategoryWithPostCount, error)
	ConvertToCategoryWithPostCountDTOs(ctx context.Context, category []*entity.Category) ([]*dto.CategoryWithPostCount, error)
}
