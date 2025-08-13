package assembler

import (
	"context"
	"dash/consts"
	"dash/model/dto"
	"dash/model/entity"
	"dash/model/property"
	"dash/service"
	"strings"
)

type BasePostAssembler interface {
	ConvertToPostOutlineDTO(ctx context.Context, post *entity.Post) (*dto.PostOutline, error)
	ConvertToPostDTO(ctx context.Context, post *entity.Post) (*dto.Post, error)
	ConvertToDetailDTO(ctx context.Context, post *entity.Post) (*dto.PostDetail, error)
}

type basePostAssemblerImpl struct {
	BasePostService service.BasePostService
	OptionService   service.OptionService
}

func NewBasePostAssembler(basePostService service.BasePostService, optionService service.OptionService) BasePostAssembler {
	return &basePostAssemblerImpl{
		BasePostService: basePostService,
		OptionService:   optionService,
	}
}

func (b *basePostAssemblerImpl) ConvertToPostOutlineDTO(ctx context.Context, post *entity.Post) (*dto.PostOutline, error) {
	postOutlineDTO := &dto.PostOutline{
		ID:         post.ID,
		Title:      post.Title,
		Status:     post.Status,
		Slug:       post.Slug,
		EditorType: post.EditorType,
		CreateTime: post.CreateTime.UnixMilli(),
	}
	if post.UpdateTime != nil {
		postOutlineDTO.UpdateTime = post.UpdateTime.UnixMilli()
	}

	fullPath := strings.Builder{}
	fullPath.WriteString("/")

	var prefix interface{}
	var err error
	if post.Type == consts.PostTypePost {
		prefix, err = b.OptionService.GetOrByDefaultWithErr(ctx, property.ArchivesPrefix, property.ArchivesPrefix.DefaultValue)
		if err != nil {
			return nil, err
		}
	} else {
		prefix, err = b.OptionService.GetOrByDefaultWithErr(ctx, property.SheetPrefix, property.SheetPrefix.DefaultValue)
		if err != nil {
			return nil, err
		}
	}
	fullPath.WriteString(prefix.(string))
	fullPath.WriteString("/")
	fullPath.WriteString(post.Slug)
	postOutlineDTO.FullPath = fullPath.String()

	return postOutlineDTO, nil
}

func (b *basePostAssemblerImpl) ConvertToPostDTO(ctx context.Context, post *entity.Post) (*dto.Post, error) {
	postOutlineDTO, err := b.ConvertToPostOutlineDTO(ctx, post)
	if err != nil {
		return nil, err
	}
	postDTO := &dto.Post{
		PostOutline: *postOutlineDTO,
		Summary:     post.Summary,
		Thumbnail:   post.Thumbnail,
		Visits:      post.Visits,
		// DisallowComment: post.DisallowComment,
		// Password:        post.Password,
		// Template:        post.Template,
		TopPriority: post.TopPriority,
		Likes:       post.Likes,
		WordCount:   post.WordCount,
		Topped:      post.TopPriority > 0,
	}
	return postDTO, nil
}

func (b *basePostAssemblerImpl) ConvertToDetailDTO(ctx context.Context, post *entity.Post) (*dto.PostDetail, error) {
	postDTO, err := b.ConvertToPostDTO(ctx, post)
	if err != nil {
		return nil, err
	}
	postDetailDTO := &dto.PostDetail{
		Post:            *postDTO,
		OriginalContent: post.OriginalContent,
		Content:         post.FormatContent,
	}
	return postDetailDTO, nil
}
