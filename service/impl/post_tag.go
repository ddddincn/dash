package impl

import (
	"context"
	"dash/consts"
	"dash/dal"
	"dash/model/dto"
	"dash/model/entity"
	"dash/model/param"
	"dash/service"

	"gorm.io/gorm"
)

type postTagServiceImpl struct {
	TagService service.TagService
	DB         *gorm.DB
}

func NewPostTagService(tagService service.TagService, db *gorm.DB) service.PostTagService {
	return &postTagServiceImpl{
		TagService: tagService,
		DB:         db,
	}
}

func (p postTagServiceImpl) ListTagsByPostID(ctx context.Context, postID int32) ([]*entity.Tag, error) {
	postTagDAL := dal.GetQueryByCtx(ctx).PostTag
	tagDAL := dal.GetQueryByCtx(ctx).Tag
	tags, err := tagDAL.WithContext(ctx).Join(&entity.PostTag{}, tagDAL.ID.EqCol(postTagDAL.TagID)).Where(postTagDAL.PostID.Eq(postID)).Find()
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (p postTagServiceImpl) ListPostsByTagID(ctx context.Context, tagID int32, status consts.PostStatus) ([]*entity.Post, error) {
	postTagDAL := dal.GetQueryByCtx(ctx).PostTag
	postDAL := dal.GetQueryByCtx(ctx).Post

	postIDsQuery := postTagDAL.WithContext(ctx).Where(postTagDAL.ID.Eq(tagID)).Select(postTagDAL.PostID)
	posts, err := postDAL.WithContext(ctx).Where(postDAL.WithContext(ctx).Columns(postDAL.ID).In(postIDsQuery), postDAL.Status.Eq(status)).Find()
	return posts, WrapDBErr(err)
}

func (p *postTagServiceImpl) ListTagMapByPostID(ctx context.Context, postIDs []int32) (map[int32][]*entity.Tag, error) {
	res := make(map[int32][]*entity.Tag, 0)
	if len(postIDs) == 0 {
		return res, nil
	}
	postTagDAL := dal.GetQueryByCtx(ctx).PostTag
	postTags, err := postTagDAL.WithContext(ctx).Where(postTagDAL.PostID.In(postIDs...)).Find()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	if len(postTags) == 0 {
		return res, nil
	}
	tagIDs := make([]int32, 0)
	for _, postTag := range postTags {
		tagIDs = append(tagIDs, postTag.TagID)
	}
	tags, err := p.TagService.ListByIDs(ctx, tagIDs)
	if err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return res, nil
	}

	tagIDMap := make(map[int32]*entity.Tag, 0)
	for _, tag := range tags {
		tagIDMap[tag.ID] = tag
	}
	for _, postTag := range postTags {
		curPostTags, ok := res[postTag.PostID]
		if !ok {
			curPostTags = make([]*entity.Tag, 0)
		}
		tag, ok := tagIDMap[postTag.TagID]
		if !ok {
			continue
		}
		curPostTags = append(curPostTags, tag)
		res[postTag.PostID] = curPostTags
	}

	return res, nil
}
func (p postTagServiceImpl) ListTagWithPostCount(ctx context.Context, sort *param.Sort) ([]*dto.TagWithPostCount, error) {
	postTagDAL := dal.GetQueryByCtx(ctx).PostTag
	tagDAL := dal.GetQueryByCtx(ctx).Tag
	tagDo := tagDAL.WithContext(ctx)

	err := BuildSort(sort, &tagDAL, &tagDo)
	if err != nil {
		return nil, err
	}

	tagWithPostCounts := make([]*struct {
		*entity.Tag
		PostCount int64 `gorm:"column:postCount"`
	}, 0)

	err = tagDo.Select(tagDAL.ALL, postTagDAL.PostID.Count().As("postCount")).LeftJoin(postTagDAL, tagDAL.ID.EqCol(postTagDAL.TagID)).Group(tagDAL.ID).Scan(&tagWithPostCounts)
	if err != nil {
		return nil, WrapDBErr(err)
	}
	tags := make([]*entity.Tag, len(tagWithPostCounts))
	for i, tagWithPostCount := range tagWithPostCounts {
		tags[i] = tagWithPostCount.Tag
	}
	tagDTOs, err := p.TagService.ConvertToTagDTOs(ctx, tags)
	if err != nil {
		return nil, err
	}
	tagDTOMap := make(map[int32]*dto.Tag, len(tagDTOs))
	for _, tagDTO := range tagDTOs {
		tagDTOMap[tagDTO.ID] = tagDTO
	}
	res := make([]*dto.TagWithPostCount, len(tagWithPostCounts))
	for i, tagWithPostCount := range tagWithPostCounts {
		tagDTO, ok := tagDTOMap[tagWithPostCount.ID]
		if !ok {
			continue
		}
		res[i] = &dto.TagWithPostCount{
			Tag:       tagDTO,
			PostCount: tagWithPostCount.PostCount,
		}
	}
	return res, nil
}
