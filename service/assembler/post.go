package assembler

import (
	"context"
	"dash/model/dto"
	"dash/model/entity"
	"dash/service"
	"sort"
	"time"

	"dash/model/vo"
)

type PostAssembler interface {
	BasePostAssembler
	ConvertToPostVOs(ctx context.Context, posts []*entity.Post) ([]*vo.Post, error)
	ConvertToDetailVO(ctx context.Context, post *entity.Post) (*vo.PostDetail, error)
	ConvertToArchivesVOs(ctx context.Context, posts []*entity.Post) ([]*vo.Archive, error)
	ConvertToCategoryVOs(ctx context.Context, posts []*entity.Post) ([]*vo.Category, error)
	ConvertToTagVOs(ctx context.Context, posts []*entity.Post) ([]*vo.Tag, error)
}

type postAssemblerImpl struct {
	BasePostAssembler
	PostService    service.PostService
	PostTagService service.PostTagService
	TagService     service.TagService
	// BaseCommentService  service.BaseCommentService
	// PostCommentService  service.PostCommentService
	PostCategoryService service.PostCategoryService
	CategoryService     service.CategoryService
	// MetaService         service.MetaService
}

func NewPostAssembler(postService service.PostService, postTagService service.PostTagService, tagService service.TagService, postCategoryService service.PostCategoryService, categoryService service.CategoryService, basePostAssembler BasePostAssembler) PostAssembler {
	return &postAssemblerImpl{
		PostService:    postService,
		PostTagService: postTagService,
		TagService:     tagService,
		// BaseCommentService:  baseCommentService,
		// PostCommentService:  postCommentService,
		PostCategoryService: postCategoryService,
		CategoryService:     categoryService,
		// MetaService:         metaService,
		BasePostAssembler: basePostAssembler,
	}
}

func (p *postAssemblerImpl) ConvertToPostVOs(ctx context.Context, posts []*entity.Post) ([]*vo.Post, error) {
	postVOs := make([]*vo.Post, 0)
	postIDs := make([]int32, 0)
	for _, post := range posts {
		postIDs = append(postIDs, post.ID)
	}

	var err error
	postTagsMap, err := p.PostTagService.ListTagMapByPostID(ctx, postIDs) // postID to []entity.tag
	if err != nil {
		return nil, err
	}
	tagDTOMap := make(map[int32]*dto.Tag) // tagID to dto.Tag
	for _, tags := range postTagsMap {
		for _, tag := range tags { // range []entity.tag
			if _, ok := tagDTOMap[tag.ID]; !ok { // if tag does not exist in tagDTOMap, insert one
				tagDTO, err := p.TagService.ConvertToTagDTO(ctx, tag)
				if err != nil {
					return nil, err
				}
				tagDTOMap[tag.ID] = tagDTO
			}
		}
	}

	postCategoryMap, err := p.PostCategoryService.ListCategoryMapByPostID(ctx, postIDs) // postID to []entity.category
	if err != nil {
		return nil, err
	}
	categoryDTOMap := make(map[int32]*dto.Category) // categoryID to dto.category
	for _, categories := range postCategoryMap {
		for _, category := range categories { // range []entity.category
			if _, ok := categoryDTOMap[category.ID]; !ok {
				categoryDTO, err_ := p.CategoryService.ConvertToCategoryDTO(ctx, category)
				if err_ != nil {
					return nil, err_
				}
				categoryDTOMap[category.ID] = categoryDTO
			}
		}
	}

	for _, post := range posts {
		postVO := &vo.Post{}

		if categories, ok := postCategoryMap[post.ID]; ok {
			categoryDTOs := make([]*dto.Category, 0)
			for _, category := range categories {
				if categoryDTO, ok := categoryDTOMap[category.ID]; ok {
					categoryDTOs = append(categoryDTOs, categoryDTO)
				}
			}
			postVO.Categories = categoryDTOs
		}

		if tags, ok := postTagsMap[post.ID]; ok {
			tagDTOs := make([]*dto.Tag, 0)
			for _, tag := range tags {
				if tagDTO, ok := tagDTOMap[tag.ID]; ok {
					tagDTOs = append(tagDTOs, tagDTO)
				}
			}
			postVO.Tags = tagDTOs
		}
		postDTO, err := p.ConvertToPostDTO(ctx, post)
		if err != nil {
			return nil, err
		}
		postVO.Post = *postDTO

		postVOs = append(postVOs, postVO)
	}

	return postVOs, nil
}

func (p *postAssemblerImpl) ConvertToDetailVO(ctx context.Context, post *entity.Post) (*vo.PostDetail, error) {
	if post == nil {
		return nil, nil
	}
	postDetailVO := &vo.PostDetail{}
	postDetailDTO, err := p.ConvertToDetailDTO(ctx, post)
	if err != nil {
		return nil, err
	}
	postDetailVO.PostDetail = *postDetailDTO

	tags, err := p.PostTagService.ListTagsByPostID(ctx, post.ID)
	if err != nil {
		return nil, err
	}
	tagDTOs, err := p.TagService.ConvertToTagDTOs(ctx, tags)
	if err != nil {
		return nil, err
	}
	postDetailVO.Tags = tagDTOs

	categories, err := p.PostCategoryService.ListCategoriesByPostID(ctx, post.ID)
	if err != nil {
		return nil, err
	}
	categoryDTOs, err := p.CategoryService.ConvertToCategoryDTOs(ctx, categories)
	if err != nil {
		return nil, err
	}
	postDetailVO.Categories = categoryDTOs

	prePost, err := p.PostService.GetPrevPosts(ctx, post, 1)
	if err != nil {
		return nil, err
	}
	nextPost, err := p.PostService.GetNextPosts(ctx, post, 1)
	if err != nil {
		return nil, err
	}
	if len(prePost) != 0 {
		postDetailVO.PrePost, err = p.BasePostAssembler.ConvertToPostOutlineDTO(ctx, prePost[0])
		if err != nil {
			return nil, err
		}
	}
	if len(nextPost) != 0 {
		postDetailVO.NextPost, err = p.BasePostAssembler.ConvertToPostOutlineDTO(ctx, nextPost[0])
		if err != nil {
			return nil, err
		}
	}

	return postDetailVO, nil
}

func (p *postAssemblerImpl) ConvertToArchivesVOs(ctx context.Context, posts []*entity.Post) ([]*vo.Archive, error) {
	postVos, err := p.ConvertToPostVOs(ctx, posts)
	if err != nil {
		return nil, err
	}
	archivesVos := make([]*vo.Archive, 0)
	archiveToPostMap := make(map[int][]*vo.Post)
	for _, postVo := range postVos {
		archiveToPostMap[time.UnixMilli(postVo.CreateTime).Year()] = append(archiveToPostMap[time.UnixMilli(postVo.CreateTime).Year()], postVo)
	}
	for year, postVos := range archiveToPostMap {
		sort.Slice(postVos, func(i, j int) bool {
			return postVos[i].CreateTime >= postVos[j].CreateTime
		})
		archivesVos = append(archivesVos, &vo.Archive{Year: year, Posts: postVos})
	}
	sort.Slice(archivesVos, func(i, j int) bool {
		return archivesVos[i].Year >= archivesVos[j].Year
	})
	return archivesVos, nil
}

func (p *postAssemblerImpl) ConvertToCategoryVOs(ctx context.Context, posts []*entity.Post) ([]*vo.Category, error) {
	// 将文章实体转换为文章VO对象
	postVos, err := p.ConvertToPostVOs(ctx, posts)
	if err != nil {
		return nil, err
	}

	// 创建分类到文章的映射
	categoryToPostsMap := make(map[string][]*vo.Post)
	categoryInfoMap := make(map[string]*dto.Category)

	// 遍历所有文章VO，按分类进行分组
	for _, postVo := range postVos {
		// 遍历文章的所有分类
		for _, category := range postVo.Categories {
			// 使用分类slug作为key进行分组
			categoryToPostsMap[category.Slug] = append(categoryToPostsMap[category.Slug], postVo)
			// 保存分类信息
			categoryInfoMap[category.Slug] = category
		}
	}

	// 创建分类VO对象列表
	categoryVos := make([]*vo.Category, 0)
	for slug, posts := range categoryToPostsMap {
		// 获取分类信息
		categoryInfo := categoryInfoMap[slug]
		// 创建分类VO对象
		categoryVo := &vo.Category{
			Category: categoryInfo,
			Posts:    posts, // 该分类下的文章列表
		}
		categoryVos = append(categoryVos, categoryVo)
	}

	return categoryVos, nil
}

func (p *postAssemblerImpl) ConvertToTagVOs(ctx context.Context, posts []*entity.Post) ([]*vo.Tag, error) {
	postVOs, err := p.ConvertToPostVOs(ctx, posts)
	if err != nil {
		return nil, err
	}

	tagToPostsMap := make(map[string][]*vo.Post)
	tagInfoMap := make(map[string]*dto.Tag)

	for _, postVO := range postVOs {
		for _, tag := range postVO.Tags {
			tagToPostsMap[tag.Slug] = append(tagToPostsMap[tag.Slug], postVO)
			tagInfoMap[tag.Slug] = tag
		}
	}

	// 创建分类VO对象列表
	tagVos := make([]*vo.Tag, 0)
	for slug, posts := range tagToPostsMap {
		// 获取分类信息
		tagInfo := tagInfoMap[slug]
		// 创建分类VO对象
		tagVo := &vo.Tag{
			Tag:   tagInfo,
			Posts: posts, // 该分类下的文章列表
		}
		tagVos = append(tagVos, tagVo)
	}

	return tagVos, nil
}
