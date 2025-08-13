package impl

import (
	"context"
	"dash/dal"
	"dash/model/dto"
	"dash/model/entity"
	"dash/model/param"
	"dash/model/property"
	"dash/service"
	"dash/utils"
	"dash/utils/xerr"
	"strings"
	"time"

	"gorm.io/gen/field"
)

type categoryServiceImpl struct {
	OptionService service.OptionService
}

func NewCategoryService(optionService service.OptionService) service.CategoryService {
	return &categoryServiceImpl{
		OptionService: optionService,
	}
}

func (c *categoryServiceImpl) Create(ctx context.Context, categoryParam *param.Category) (*entity.Category, error) {
	if categoryParam.Slug == "" { // correct parameter,slug may be empty
		categoryParam.Slug = utils.Slug(categoryParam.Name)
	} else {
		categoryParam.Slug = utils.Slug(categoryParam.Slug)
	}

	categoryDAL := dal.GetQueryByCtx(ctx).Category
	// determine if name and slug exists
	count, err := categoryDAL.WithContext(ctx).
		Where(
			field.Or(
				categoryDAL.Name.Eq(categoryParam.Name),
				categoryDAL.Slug.Eq(categoryParam.Slug),
			),
		).Count()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	if count > 0 {
		return nil, xerr.BadParam.New("invalid parameter").WithMsg("category name or slug has existed already").WithStatus(xerr.StatusBadRequest)
	}

	category := &entity.Category{
		CreateTime:  time.Now(),
		Name:        categoryParam.Name,
		Slug:        categoryParam.Slug,
		Description: categoryParam.Description,
		Thumbnail:   categoryParam.Thumbnail,
		Priority:    categoryParam.Priority,
	}
	err = categoryDAL.WithContext(ctx).Create(category)
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return category, nil
}

func (c *categoryServiceImpl) DeleteByID(ctx context.Context, id int32) error {
	err := dal.Transaction(ctx, func(txCtx context.Context) error {
		categoryDAL := dal.GetQueryByCtx(ctx).Category // delete info from category table, 1 to 1
		_, err := categoryDAL.WithContext(txCtx).Where(categoryDAL.ID.Value(id)).Delete()
		if err != nil {
			return WrapDBErr(err)
		}

		postCategoryDAL := dal.GetQueryByCtx(ctx).PostCategory // delete info from post_category table, 1 to n
		_, err = postCategoryDAL.WithContext(txCtx).Where(postCategoryDAL.CategoryID.Eq(id)).Delete()
		if err != nil {
			return WrapDBErr(err)
		}
		return nil
	})
	return err
}

func (c *categoryServiceImpl) UpdateByID(ctx context.Context, id int32, categoryParam *param.Category) (*entity.Category, error) {
	if categoryParam.Slug == "" { // correct parameter,slug may be empty
		categoryParam.Slug = utils.Slug(categoryParam.Name)
	} else {
		categoryParam.Slug = utils.Slug(categoryParam.Slug)
	}

	categoryDAL := dal.GetQueryByCtx(ctx).Category
	// determine if category exist
	_, err := categoryDAL.WithContext(ctx).Where(categoryDAL.ID.Eq(id)).First()
	if err != nil {
		return nil, WrapDBErr(err)
	}

	// check for records with same name or slug
	count, err := categoryDAL.WithContext(ctx).
		Where(categoryDAL.ID.Neq(id)).
		Where(
			field.Or(
				categoryDAL.Name.Eq(categoryParam.Name),
				categoryDAL.Slug.Eq(categoryParam.Slug),
			),
		).
		Count()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	if count > 0 {
		return nil, xerr.BadParam.New("invalid parameter").WithMsg("category name or slug has existed already").WithStatus(xerr.StatusBadRequest)
	}

	// update record
	updateResult, err := categoryDAL.WithContext(ctx).Where(categoryDAL.ID.Eq(id)).UpdateSimple(
		categoryDAL.UpdateTime.Value(time.Now()),
		categoryDAL.Name.Value(categoryParam.Name),
		categoryDAL.Slug.Value(categoryParam.Slug),
		categoryDAL.Description.Value(categoryParam.Description),
		categoryDAL.Thumbnail.Value(categoryParam.Thumbnail),
		categoryDAL.Priority.Value(categoryParam.Priority),
	)
	if err != nil {
		return nil, WrapDBErr(err)
	}
	if updateResult.RowsAffected != 1 {
		return nil, xerr.NoType.New("update category failed id=%v", id).WithStatus(xerr.StatusInternalServerError).WithMsg("update tag failed")
	}
	category, err := categoryDAL.WithContext(ctx).Where(categoryDAL.ID.Value(id)).First()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return category, err
}

func (c *categoryServiceImpl) List(ctx context.Context, sort *param.Sort) ([]*entity.Category, error) {
	categoryDAL := dal.GetQueryByCtx(ctx).Category
	categoryDO := categoryDAL.WithContext(ctx)
	err := BuildSort(sort, &categoryDAL, &categoryDO)
	if err != nil {
		return nil, err
	}
	categories, err := categoryDO.Find()
	return categories, err
}

func (c *categoryServiceImpl) ListByIDs(ctx context.Context, ids []int32) ([]*entity.Category, error) {
	categoryDAL := dal.GetQueryByCtx(ctx).Category
	categories, err := categoryDAL.WithContext(ctx).Where(categoryDAL.ID.In(ids...)).Find()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return categories, nil
}

func (c *categoryServiceImpl) GetCategoryByID(ctx context.Context, id int32) (*entity.Category, error) {
	categoryDAL := dal.GetQueryByCtx(ctx).Category
	category, err := categoryDAL.WithContext(ctx).Where(categoryDAL.ID.Eq(id)).First()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return category, nil
}

func (c *categoryServiceImpl) GetCategoryByName(ctx context.Context, name string) (*entity.Category, error) {
	categoryDAL := dal.GetQueryByCtx(ctx).Category
	category, err := categoryDAL.WithContext(ctx).Where(categoryDAL.Name.Eq(name)).First()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return category, nil
}

func (c *categoryServiceImpl) GetCategoryBySlug(ctx context.Context, slug string) (*entity.Category, error) {
	categoryDAL := dal.GetQueryByCtx(ctx).Category
	category, err := categoryDAL.WithContext(ctx).Where(categoryDAL.Slug.Eq(slug)).First()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return category, nil
}

func (c categoryServiceImpl) GetCategoriesCount(ctx context.Context) (int64, error) {
	categoryDAL := dal.GetQueryByCtx(ctx).Category
	count, err := categoryDAL.WithContext(ctx).Count()
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (c *categoryServiceImpl) ConvertToCategoryDTO(ctx context.Context, category *entity.Category) (*dto.Category, error) {
	categoryDTO := &dto.Category{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		Thumbnail:   category.Thumbnail,
		CreateTime:  category.CreateTime.UnixMilli(),
		Priority:    category.Priority,
	}

	fullPath := strings.Builder{}

	fullPath.WriteString("/")
	categoryPrefix, err := c.OptionService.GetOrByDefaultWithErr(ctx, property.CategoriesPrefix, property.CategoriesPrefix.DefaultValue)
	if err != nil {
		return nil, err
	}
	fullPath.WriteString(categoryPrefix.(string))
	fullPath.WriteString("/")
	fullPath.WriteString(category.Slug)

	categoryDTO.FullPath = fullPath.String()
	return categoryDTO, nil
}

func (c *categoryServiceImpl) ConvertToCategoryDTOs(ctx context.Context, categories []*entity.Category) ([]*dto.Category, error) {
	result := make([]*dto.Category, len(categories))

	categoryPrefix, err := c.OptionService.GetOrByDefaultWithErr(ctx, property.CategoriesPrefix, property.CategoriesPrefix.DefaultValue)
	if err != nil {
		return nil, err
	}

	for i, category := range categories {
		categoryDTO := &dto.Category{}
		categoryDTO.ID = category.ID
		categoryDTO.Thumbnail = category.Thumbnail
		// categoryDTO.ParentID = category.ParentID
		categoryDTO.Name = category.Name
		categoryDTO.CreateTime = category.CreateTime.UnixMilli()
		categoryDTO.Description = category.Description
		categoryDTO.Slug = category.Slug
		categoryDTO.Priority = category.Priority

		fullPath := strings.Builder{}
		fullPath.WriteString("/")
		fullPath.WriteString(categoryPrefix.(string))
		fullPath.WriteString("/")
		fullPath.WriteString(category.Slug)
		categoryDTO.FullPath = fullPath.String()
		result[i] = categoryDTO
	}
	return result, nil
}

func (c *categoryServiceImpl) ConvertToCategoryWithPostCountDTO(ctx context.Context, category *entity.Category) (*dto.CategoryWithPostCount, error) {
	categoryDTO, err := c.ConvertToCategoryDTO(ctx, category)
	if err != nil {
		return nil, err
	}
	postCategoryDAL := dal.GetQueryByCtx(ctx).PostCategory
	count, err := postCategoryDAL.WithContext(ctx).Where(postCategoryDAL.CategoryID.Eq(categoryDTO.ID)).Count()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	categoryWithPostCountDTO := &dto.CategoryWithPostCount{
		Category:  categoryDTO,
		PostCount: count,
	}
	return categoryWithPostCountDTO, nil
}

func (c *categoryServiceImpl) ConvertToCategoryWithPostCountDTOs(ctx context.Context, categories []*entity.Category) ([]*dto.CategoryWithPostCount, error) {
	categoryDTOs, err := c.ConvertToCategoryDTOs(ctx, categories)
	if err != nil {
		return nil, err
	}
	categoryWithPostCountDTOs := make([]*dto.CategoryWithPostCount, 0)
	postCategoryDAL := dal.GetQueryByCtx(ctx).PostCategory
	for _, categoryDTO := range categoryDTOs {
		count, err := postCategoryDAL.WithContext(ctx).Where(postCategoryDAL.CategoryID.Eq(categoryDTO.ID)).Count()
		if err != nil {
			return nil, WrapDBErr(err)
		}
		categoryWithPostCountDTO := &dto.CategoryWithPostCount{
			Category:  categoryDTO,
			PostCount: count,
		}
		categoryWithPostCountDTOs = append(categoryWithPostCountDTOs, categoryWithPostCountDTO)
	}
	return categoryWithPostCountDTOs, nil
}
