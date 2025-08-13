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

type TagHandler struct {
	OptionService  service.OptionService
	TagService     service.TagService
	PostService    service.PostService
	PostTagService service.PostTagService
	PostAssembler  assembler.PostAssembler
}

func NewTagHandler(optionService service.OptionService, tagService service.TagService, postService service.PostService, postTagService service.PostTagService, postAssembler assembler.PostAssembler) *TagHandler {
	return &TagHandler{
		OptionService:  optionService,
		TagService:     tagService,
		PostService:    postService,
		PostTagService: postTagService,
		PostAssembler:  postAssembler,
	}
}

func (t *TagHandler) ListTags(ctx *gin.Context) (interface{}, error) {
	tagQuery := struct {
		*param.Sort
		Detail *bool `json:"detail" form:"detail"`
	}{}
	err := ctx.ShouldBindQuery(&tagQuery)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(e.Error())
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest)
	}
	if tagQuery.Sort == nil || len(tagQuery.Sort.Fields) == 0 {
		tagQuery.Sort = &param.Sort{Fields: []string{"create_time,asc"}}
	}
	tags, err := t.TagService.List(ctx, tagQuery.Sort)
	if err != nil {
		return nil, err
	}
	if tagQuery.Detail != nil && *tagQuery.Detail {
		return t.TagService.ConvertToTagWithPostCountDTOs(ctx, tags)
	}

	return t.TagService.ConvertToTagDTOs(ctx, tags)
}

func (t *TagHandler) ListTagsWithPosts(ctx *gin.Context) (interface{}, error) {
	sort := &param.Sort{
		Fields: []string{"create_time,asc"},
	}
	tags, err := t.TagService.List(ctx, sort)
	if err != nil {
		return nil, err
	}
	tagDTOs, err := t.TagService.ConvertToTagDTOs(ctx, tags)
	if err != nil {
		return nil, err
	}
	tagVOs := make([]*vo.Tag, 0)
	for _, tagDTO := range tagDTOs {
		tagVO := &vo.Tag{}
		posts, err := t.PostTagService.ListPostsByTagID(ctx, tagDTO.ID, consts.PostStatusPublished)
		if err != nil {
			return nil, err
		}
		postVOs, err := t.PostAssembler.ConvertToPostVOs(ctx, posts)
		if err != nil {
			return nil, err
		}
		tagVO.Tag = tagDTO
		tagVO.Posts = postVOs
		tagVOs = append(tagVOs, tagVO)
	}

	return tagVOs, nil
}

func (t *TagHandler) GetTagByID(ctx *gin.Context) (interface{}, error) {
	id, err := utils.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	tag, err := t.TagService.GetTagByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return t.TagService.ConvertToTagDTO(ctx, tag)
}

func (t *TagHandler) CreateTag(ctx *gin.Context) (interface{}, error) {
	tagParam := &param.Tag{}
	err := ctx.ShouldBindJSON(&tagParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(e.Error())
		}
		return nil, err
	}
	tag, err := t.TagService.Create(ctx, tagParam)
	if err != nil {
		return nil, err
	}
	return t.TagService.ConvertToTagDTO(ctx, tag)
}

func (t *TagHandler) UpdateTag(ctx *gin.Context) (interface{}, error) {
	tagParam := &param.Tag{}
	err := ctx.ShouldBindJSON(&tagParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(e.Error())
		}
		return nil, err
	}
	id, err := utils.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	tag, err := t.TagService.UpdateByID(ctx, id, tagParam)
	if err != nil {
		return nil, err
	}
	return t.TagService.ConvertToTagDTO(ctx, tag)
}

func (t *TagHandler) DeleteTag(ctx *gin.Context) (interface{}, error) {
	id, err := utils.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	return nil, t.TagService.DeleteByID(ctx, id)
}

func (t *TagHandler) ListPostsByTagSlug(ctx *gin.Context) (interface{}, error) {
	slug, err := utils.ParamString(ctx, "slug")
	if err != nil {
		return nil, err
	}
	page, err := utils.MustGetQueryInt32(ctx, "page")
	if err != nil {
		return nil, err
	}
	pageSize := t.OptionService.GetOrByDefault(ctx, property.TagPageSize).(int)
	tag, err := t.TagService.GetTagBySlug(ctx, slug)
	if err != nil {
		return "", err
	}
	id := tag.ID
	pageQuery := param.PostQuery{
		Page: param.Page{
			PageNum:  int(page),
			PageSize: pageSize,
		},
		TagID: &id,
		Sort: &param.Sort{
			Fields: []string{"create_time,desc"},
		},
		Statuses: []*consts.PostStatus{consts.PostStatusPublished.Ptr()},
	}
	posts, totalPage, err := t.PostService.Page(ctx, pageQuery)
	if err != nil {
		return "", err
	}
	tagVOs, err := t.PostAssembler.ConvertToTagVOs(ctx, posts)
	if err != nil {
		return "", err
	}
	postPage := dto.NewPage(tagVOs, totalPage, param.Page{
		PageNum:  int(page),
		PageSize: pageSize,
	})
	return postPage, nil
}

// func (t *TagHandler) Tags(ctx *gin.Context) (interface{}, error) {
// 	tags, err := t.TagService.List(ctx, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	tagVOs := make([]*vo.Tag, 0)

// 	for _, tag := range tags {
// 		postQuery := param.PostQuery{
// 			Page: param.Page{
// 				PageNum:  0,
// 				PageSize: 15, // 限制每个分类最多返回15篇文章
// 			},
// 			TagID:    &tag.ID,
// 			Statuses: []*consts.PostStatus{consts.PostStatusPublished.Ptr()},
// 		}

// 		posts, _, err := t.PostService.Page(ctx, postQuery)
// 		if err != nil {
// 			return nil, err
// 		}
// 		postVOs, err := t.PostAssembler.ConvertToListVO(ctx, posts)
// 		if err != nil {
// 			return nil, err
// 		}
// 		tagVO := &vo.Tag{
// 			Name:  tag.Name,
// 			Slug:  tag.Slug,
// 			Posts: postVOs,
// 		}
// 		tagVOs = append(tagVOs, tagVO)
// 	}
// 	return tagVOs, nil
// }

// func (t *TagHandler) Tag(ctx *gin.Context) (interface{}, error) {
// 	slug, err := utils.ParamString(ctx, "slug")
// 	if err != nil {
// 		return nil, err
// 	}
// 	return t.PostPresenter.Tag(ctx, slug, 0)
// }

// func (t *TagHandler) TagPage(ctx *gin.Context) (interface{}, error) {
// 	slug, err := utils.ParamString(ctx, "slug")
// 	if err != nil {
// 		return nil, err
// 	}
// 	page, err := utils.ParamInt32(ctx, "page")
// 	if err != nil {
// 		return nil, err
// 	}
// 	return t.PostPresenter.Tag(ctx, slug, int(page)-1)
// }

// func (t *TagHandler) ListTags(ctx *gin.Context) (interface{}, error) {
// 	sort := param.Sort{}
// 	err := ctx.ShouldBindQuery(&sort)
// 	if err != nil {
// 		return nil, xerr.WithMsg(err, "sort parameter error").WithStatus(xerr.StatusBadRequest)
// 	}
// 	if len(sort.Fields) == 0 {
// 		sort.Fields = append(sort.Fields, "createTime,desc")
// 	}
// 	tags, err := t.TagService.List(ctx, &sort)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return t.TagService.ConvertToTagDTOs(ctx, tags)
// }

// func (t *TagHandler) CreateTag(ctx *gin.Context) (interface{}, error) {
// 	tagParam := &param.Tag{}
// 	err := ctx.ShouldBindJSON(tagParam)
// 	if err != nil {
// 		// e := validator.ValidationErrors{}
// 		// if errors.As(err, &e) {
// 		// 	return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
// 		// }
// 		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
// 	}
// 	tag, err := t.TagService.Create(ctx, tagParam)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return t.TagService.ConvertToTagDTO(ctx, tag)
// }
