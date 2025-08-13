package handler

import (
	"dash/consts"
	"dash/model/dto"
	"dash/model/property"
	"dash/service"
	"time"

	"github.com/gin-gonic/gin"
)

type StatisticsHandler struct {
	PostService     service.PostService
	TagService      service.TagService
	CategoryService service.CategoryService
	OptionService   service.OptionService
}

func NewStatisticsHandler(postService service.PostService, tagService service.TagService, categoryService service.CategoryService, optionService service.OptionService) *StatisticsHandler {
	return &StatisticsHandler{
		PostService:     postService,
		TagService:      tagService,
		CategoryService: categoryService,
		OptionService:   optionService,
	}
}

func (s *StatisticsHandler) Statistic(ctx *gin.Context) (interface{}, error) {
	var statistic dto.Statistic
	postCount, err := s.PostService.GetPostCountByStatus(ctx, consts.PostStatusPublished)
	if err != nil {
		return nil, err
	}
	tagCount, err := s.TagService.GetTagsCount(ctx)
	if err != nil {
		return nil, err
	}
	categoryCount, err := s.CategoryService.GetCategoriesCount(ctx)
	if err != nil {
		return nil, err
	}
	postVisitCount, err := s.PostService.GetVisitCount(ctx)
	if err != nil {
		return nil, err
	}
	postLikeCount, err := s.PostService.GetLikeCount(ctx)
	if err != nil {
		return nil, err
	}
	birthday, err := s.OptionService.GetOrByDefaultWithErr(ctx, property.BirthDay, time.Now().UnixMilli())
	if err != nil {
		return nil, err
	}
	statistic.PostCount = postCount
	statistic.TagCount = tagCount
	statistic.CategoryCount = categoryCount
	statistic.VisitCount = postVisitCount
	statistic.LikeCount = postLikeCount
	statistic.Birthday = birthday.(int64)
	statistic.EstablishDays = (time.Now().UnixMilli() - birthday.(int64)) / (1000 * 24 * 3600)
	return &statistic, nil
}
