package handler

import (
	"dash/consts"
	"dash/model/dto"
	"dash/model/param"
	"dash/model/property"
	"dash/model/vo"
	"dash/service"
	"dash/service/assembler"
	"dash/utils"
	"dash/utils/xerr"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type CategoryHandler struct {
	OptionService       service.OptionService
	CategoryService     service.CategoryService
	PostService         service.PostService
	PostCategoryService service.PostCategoryService
	PostAssembler       assembler.PostAssembler
}

func NewCategoryHandler(optionService service.OptionService, categoryService service.CategoryService, postService service.PostService, postCategoryService service.PostCategoryService, postAssembler assembler.PostAssembler) *CategoryHandler {
	return &CategoryHandler{
		OptionService:       optionService,
		CategoryService:     categoryService,
		PostService:         postService,
		PostCategoryService: postCategoryService,
		PostAssembler:       postAssembler,
	}
}

func (c *CategoryHandler) ListCategories(ctx *gin.Context) (interface{}, error) {
	categoryQuery := struct {
		*param.Sort
		Detail *bool `json:"detail" form:"detail"`
	}{}

	err := ctx.ShouldBindQuery(&categoryQuery)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	if categoryQuery.Sort == nil || len(categoryQuery.Sort.Fields) == 0 {
		categoryQuery.Sort = &param.Sort{Fields: []string{"priority,asc"}}
	}
	categories, err := c.CategoryService.List(ctx, categoryQuery.Sort)
	if err != nil {
		return nil, err
	}
	if categoryQuery.Detail != nil && *categoryQuery.Detail {
		return c.CategoryService.ConvertToCategoryWithPostCountDTOs(ctx, categories)
	}

	return c.CategoryService.ConvertToCategoryDTOs(ctx, categories)
}

func (c *CategoryHandler) ListCategoriesWithPosts(ctx *gin.Context) (interface{}, error) {
	sort := &param.Sort{
		Fields: []string{"create_time,asc"},
	}
	categories, err := c.CategoryService.List(ctx, sort)
	if err != nil {
		return nil, err
	}
	categoryDTOs, err := c.CategoryService.ConvertToCategoryDTOs(ctx, categories)
	if err != nil {
		return nil, err
	}
	categoryVOs := make([]*vo.Category, 0)
	for _, categoryDTO := range categoryDTOs {
		categoryVO := &vo.Category{}
		posts, err := c.PostCategoryService.ListPostsByCategoryID(ctx, categoryDTO.ID, consts.PostStatusPublished)
		if err != nil {
			return nil, err
		}
		postVOs, err := c.PostAssembler.ConvertToPostVOs(ctx, posts)
		if err != nil {
			return nil, err
		}
		categoryVO.Category = categoryDTO
		categoryVO.Posts = postVOs
		categoryVOs = append(categoryVOs, categoryVO)
	}

	return categoryVOs, nil
}

func (c *CategoryHandler) ListPostsByCategorySlug(ctx *gin.Context) (interface{}, error) {
	slug, err := utils.ParamString(ctx, "slug")
	if err != nil {
		return nil, err
	}
	page, err := utils.MustGetQueryInt32(ctx, "page")
	if err != nil {
		return nil, err
	}
	pageSize := c.OptionService.GetOrByDefault(ctx, property.CategoryPageSize).(int)
	category, err := c.CategoryService.GetCategoryBySlug(ctx, slug)
	if err != nil {
		return "", err
	}
	id := category.ID
	pageQuery := param.PostQuery{
		Page: param.Page{
			PageNum:  int(page),
			PageSize: pageSize,
		},
		CategoryID: &id,
		Sort: &param.Sort{
			Fields: []string{"create_time,desc"},
		},
		Statuses: []*consts.PostStatus{consts.PostStatusPublished.Ptr()},
	}
	posts, totalPage, err := c.PostService.Page(ctx, pageQuery)
	if err != nil {
		return "", err
	}
	categoryVOs, err := c.PostAssembler.ConvertToCategoryVOs(ctx, posts)
	if err != nil {
		return "", err
	}
	postPage := dto.NewPage(categoryVOs, totalPage, param.Page{
		PageNum:  int(page),
		PageSize: pageSize,
	})
	return postPage, nil
}

func (c *CategoryHandler) GetCategoryByID(ctx *gin.Context) (interface{}, error) {
	id, err := utils.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	category, err := c.CategoryService.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return c.CategoryService.ConvertToCategoryDTO(ctx, category)
}

func (c *CategoryHandler) CreateCategory(ctx *gin.Context) (interface{}, error) {
	var categoryParam = &param.Category{}
	err := ctx.ShouldBindJSON(&categoryParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(e.Error())
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest)
	}
	category, err := c.CategoryService.Create(ctx, categoryParam)
	if err != nil {
		return nil, err
	}
	return c.CategoryService.ConvertToCategoryDTO(ctx, category)
}

func (c *CategoryHandler) UpdateCategory(ctx *gin.Context) (interface{}, error) {
	var categoryParam = &param.Category{}
	err := ctx.ShouldBindJSON(&categoryParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(e.Error())
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest)
	}
	categoryID, err := utils.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	category, err := c.CategoryService.UpdateByID(ctx, categoryID, categoryParam)
	if err != nil {
		return nil, err
	}
	return c.CategoryService.ConvertToCategoryDTO(ctx, category)
}

func (c *CategoryHandler) DeleteCategory(ctx *gin.Context) (interface{}, error) {
	categoryID, err := utils.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	return nil, c.CategoryService.DeleteByID(ctx, categoryID)
}
