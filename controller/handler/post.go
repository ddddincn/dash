package handler

import (
	"dash/consts"
	"dash/controller/binding"
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

type PostHandler struct {
	OptionService service.OptionService
	PostService   service.PostService
	PostAssembler assembler.PostAssembler
}

func NewPostHandler(optionService service.OptionService, postService service.PostService, postAssembler assembler.PostAssembler) *PostHandler {
	return &PostHandler{
		OptionService: optionService,
		PostService:   postService,
		PostAssembler: postAssembler,
	}
}

func (p *PostHandler) ListPosts(ctx *gin.Context) (interface{}, error) {
	postQuery := param.PostQuery{}
	err := ctx.ShouldBindWith(&postQuery, binding.CustomFormBinding)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("invalid parameter")
	}

	if postQuery.PageSize > 50 {
		postQuery.PageSize = 50
	}
	if postQuery.Sort == nil {
		postQuery.Sort = &param.Sort{Fields: []string{"top_priority,desc", "create_time,desc"}}
	}
	posts, totalCount, err := p.PostService.Page(ctx, postQuery)
	if err != nil {
		return nil, err
	}
	if postQuery.Detail != nil && *postQuery.Detail {
		postVOs, err := p.PostAssembler.ConvertToPostVOs(ctx, posts)
		return dto.NewPage(postVOs, totalCount, postQuery.Page), err
	}
	postOutlineDTOs := make([]*dto.PostOutline, 0)
	for _, post := range posts {
		postOutlineDTO, err := p.PostAssembler.ConvertToPostOutlineDTO(ctx, post)
		if err != nil {
			return nil, err
		}
		postOutlineDTOs = append(postOutlineDTOs, postOutlineDTO)
	}
	return dto.NewPage(postOutlineDTOs, totalCount, postQuery.Page), err
}

func (p *PostHandler) GetPostByID(ctx *gin.Context) (interface{}, error) {
	postID, err := utils.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	post, err := p.PostService.GetPostByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	postDetailDTO, err := p.PostAssembler.ConvertToDetailDTO(ctx, post)
	if err != nil {
		return nil, err
	}
	return postDetailDTO, nil
}

func (p *PostHandler) GetPostBySlug(ctx *gin.Context) (interface{}, error) {
	slug, err := utils.ParamString(ctx, "slug")
	if err != nil {
		return nil, err
	}
	post, err := p.PostService.GetPostBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	postDetailDTO, err := p.PostAssembler.ConvertToDetailVO(ctx, post)
	if err != nil {
		return nil, err
	}
	return postDetailDTO, nil
}

func (p *PostHandler) SearchPost(ctx *gin.Context) (interface{}, error) {
	keyword, err := utils.MustGetQueryString(ctx, "keyword")
	if err != nil {
		return nil, err
	}
	sort := param.Sort{}
	err = ctx.ShouldBindWith(&sort, binding.CustomFormBinding)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if len(sort.Fields) == 0 {
		sort = p.OptionService.GetPostSort(ctx)
	}
	// defaultPageSize := p.OptionService.GetIndexPageSize(ctx)
	page := param.Page{
		PageNum:  0,
		PageSize: 100,
	}
	postQuery := param.PostQuery{
		Page:     page,
		Sort:     &sort,
		Keyword:  &keyword,
		Statuses: []*consts.PostStatus{consts.PostStatusPublished.Ptr()},
	}
	posts, total, err := p.PostService.Page(ctx, postQuery)
	if err != nil {
		return nil, err
	}
	postVOs, err := p.PostAssembler.ConvertToPostVOs(ctx, posts)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(postVOs, total, page), nil
}

func (p *PostHandler) CreatePost(ctx *gin.Context) (interface{}, error) {
	postParam := &param.Post{}
	err := ctx.ShouldBindJSON(&postParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(e.Error())
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}

	post, err := p.PostService.Create(ctx, postParam, consts.PostTypePost)
	if err != nil {
		return nil, err
	}
	return p.PostAssembler.ConvertToPostOutlineDTO(ctx, post)
}

func (p *PostHandler) UpdatePost(ctx *gin.Context) (interface{}, error) {
	postParam := &param.Post{}
	err := ctx.ShouldBindJSON(&postParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(e.Error())
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}

	postID, err := utils.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}

	post, err := p.PostService.UpdateByID(ctx, postID, postParam, consts.PostTypePost)
	if err != nil {
		return nil, err
	}
	return p.PostAssembler.ConvertToPostOutlineDTO(ctx, post)
}

func (p *PostHandler) UpdatePostStatus(ctx *gin.Context) (interface{}, error) {
	postID, err := utils.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	statusStr, err := utils.ParamString(ctx, "status")
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	status, err := consts.PostStatusFromString(statusStr)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	if int32(status) < int32(consts.PostStatusPublished) || int32(status) > int32(consts.PostStatusIntimate) {
		return nil, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg("status error")
	}
	post, err := p.PostService.UpdateStatusByID(ctx, postID, status)
	if err != nil {
		return nil, err
	}
	return p.PostAssembler.ConvertToPostOutlineDTO(ctx, post)
}

func (p *PostHandler) UpdatePostStatusBatch(ctx *gin.Context) (interface{}, error) {
	statusStr, err := utils.ParamString(ctx, "status")
	if err != nil {
		return nil, err
	}
	status, err := consts.PostStatusFromString(statusStr)
	if err != nil {
		return nil, err
	}
	if int32(status) < int32(consts.PostStatusPublished) || int32(status) > int32(consts.PostStatusIntimate) {
		return nil, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg("status error")
	}
	ids := make([]int32, 0)
	err = ctx.ShouldBind(&ids)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("post ids error")
	}
	posts, err := p.PostService.UpdateStatusBatch(ctx, ids, status)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("unknown err")
	}
	postDTOs := make([]*dto.PostOutline, 0)
	for _, post := range posts {
		postDTO, err := p.PostAssembler.ConvertToPostOutlineDTO(ctx, post)
		if err != nil {
			return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("unknown err")
		}
		postDTOs = append(postDTOs, postDTO)
	}

	return postDTOs, nil
}

func (p *PostHandler) DeletePost(ctx *gin.Context) (interface{}, error) {
	postID, err := utils.ParamInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	return nil, p.PostService.DeleteByID(ctx, postID)
}

func (p *PostHandler) DeletePostBatch(ctx *gin.Context) (interface{}, error) {
	postIDs := make([]int32, 0)
	err := ctx.ShouldBind(&postIDs)
	if err != nil {
		return nil, xerr.WithMsg(err, "postIDs error").WithStatus(xerr.StatusBadRequest)
	}
	return nil, p.PostService.DeleteBatchByID(ctx, postIDs)
}

func (p *PostHandler) GetPostArchive(ctx *gin.Context) (interface{}, error) {
	page, err := utils.MustGetQueryInt32(ctx, "page")
	if err != nil {
		return nil, err
	}
	pageSize := p.OptionService.GetOrByDefault(ctx, property.ArchivePageSize).(int)
	postQuery := param.PostQuery{
		Page: param.Page{
			PageNum:  int(page),
			PageSize: pageSize,
		},
		Sort: &param.Sort{
			Fields: []string{"create_time,desc"},
		},
		Statuses: []*consts.PostStatus{consts.PostStatusPublished.Ptr()},
	}
	posts, totalPage, err := p.PostService.Page(ctx, postQuery)
	if err != nil {
		return "", err
	}
	archiveVOs, err := p.PostAssembler.ConvertToArchivesVOs(ctx, posts)
	if err != nil {
		return "", err
	}
	postPage := dto.NewPage(archiveVOs, totalPage, param.Page{
		PageNum:  int(page),
		PageSize: pageSize,
	})

	archivesVO := &vo.Archives{
		Page: postPage,
	}
	return archivesVO, nil
}
