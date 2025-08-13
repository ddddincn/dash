package impl

import (
	"context"
	"dash/consts"
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
	"gorm.io/gorm"
)

type tagServiceImpl struct {
	OptionService service.OptionService
	DB            *gorm.DB
}

func NewTagService(optionService service.OptionService, db *gorm.DB) service.TagService {
	return &tagServiceImpl{
		OptionService: optionService,
		DB:            db,
	}
}

func (t *tagServiceImpl) Create(ctx context.Context, tagParam *param.Tag) (*entity.Tag, error) {
	if tagParam.Slug == "" { // correct parameter,slug and color may be empty
		tagParam.Slug = utils.Slug(tagParam.Name)
	} else {
		tagParam.Slug = utils.Slug(tagParam.Slug)
	}
	if tagParam.Color == "" {
		tagParam.Color = consts.DashDefaultTagColor
	}

	tagDAL := dal.GetQueryByCtx(ctx).Tag
	// determine if name and slug exists
	count, err := tagDAL.WithContext(ctx).
		Where(
			field.Or(
				tagDAL.Name.Eq(tagParam.Name),
				tagDAL.Slug.Eq(tagParam.Slug),
			),
		).Count()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	if count > 0 {
		return nil, xerr.BadParam.New("invalid parameter").WithMsg("tag name or slug has exist already").WithStatus(xerr.StatusBadRequest)
	}

	tag := &entity.Tag{
		CreateTime: time.Now(),
		Name:       tagParam.Name,
		Slug:       tagParam.Slug,
		Thumbnail:  tagParam.Thumbnail,
		Color:      tagParam.Color,
	}
	err = tagDAL.WithContext(ctx).Create(tag)
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return tag, nil
}

func (t *tagServiceImpl) DeleteByID(ctx context.Context, id int32) error {
	err := dal.Transaction(ctx, func(txCtx context.Context) error {
		tagDAL := dal.GetQueryByCtx(ctx).Tag
		_, err := tagDAL.WithContext(txCtx).Where(tagDAL.ID.Value(id)).Delete()
		if err != nil {
			return WrapDBErr(err)
		}

		postTagDAL := dal.GetQueryByCtx(ctx).PostTag
		_, err = postTagDAL.WithContext(txCtx).Where(postTagDAL.TagID.Eq(id)).Delete()
		if err != nil {
			return WrapDBErr(err)
		}
		return nil
	})
	return err
}

func (t *tagServiceImpl) UpdateByID(ctx context.Context, id int32, tagParam *param.Tag) (*entity.Tag, error) {
	if tagParam.Slug == "" {
		tagParam.Slug = utils.Slug(tagParam.Name)
	} else {
		tagParam.Slug = utils.Slug(tagParam.Slug)
	}
	if tagParam.Color == "" {
		tagParam.Color = consts.DashDefaultTagColor
	}

	tagDAL := dal.GetQueryByCtx(ctx).Tag
	// determine if tag exist
	_, err := tagDAL.WithContext(ctx).Where(tagDAL.ID.Eq(id)).First()
	if err != nil {
		return nil, WrapDBErr(err)
	}

	count, err := tagDAL.WithContext(ctx).
		Where(tagDAL.ID.Neq(id)).
		Where(
			field.Or(
				tagDAL.Name.Eq(tagParam.Name),
				tagDAL.Slug.Eq(tagParam.Slug),
			),
		).Count()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	if count > 0 {
		return nil, xerr.BadParam.New("invalid parameter").WithMsg("tag name or slug has exist already").WithStatus(xerr.StatusBadRequest)
	}

	updateResult, err := tagDAL.WithContext(ctx).Where(tagDAL.ID.Eq(id)).UpdateSimple(
		tagDAL.UpdateTime.Value(time.Now()),
		tagDAL.Name.Value(tagParam.Name),
		tagDAL.Slug.Value(tagParam.Slug),
		tagDAL.Thumbnail.Value(tagParam.Thumbnail),
		tagDAL.Color.Value(tagParam.Color),
	)
	if err != nil {
		return nil, WrapDBErr(err)
	}
	if updateResult.RowsAffected != 1 {
		return nil, xerr.NoType.New("update tag failed id=%v", id).WithStatus(xerr.StatusInternalServerError).WithMsg("update tag failed")
	}

	tag, err := tagDAL.WithContext(ctx).Where(tagDAL.ID.Value(id)).First()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return tag, nil
}

func (t *tagServiceImpl) List(ctx context.Context, sort *param.Sort) ([]*entity.Tag, error) {
	tagDAL := dal.GetQueryByCtx(ctx).Tag
	tagDO := tagDAL.WithContext(ctx)
	err := BuildSort(sort, &tagDAL, &tagDO)
	if err != nil {
		return nil, err
	}
	tags, err := tagDAL.WithContext(ctx).Where().Find()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return tags, nil
}

func (t *tagServiceImpl) ListByIDs(ctx context.Context, tagIDs []int32) ([]*entity.Tag, error) {
	if len(tagIDs) == 0 {
		return make([]*entity.Tag, 0), nil
	}
	tagDAL := dal.GetQueryByCtx(ctx).Tag
	tags, err := tagDAL.WithContext(ctx).Where(tagDAL.ID.In(tagIDs...)).Find()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return tags, nil
}

func (t *tagServiceImpl) GetTagByID(ctx context.Context, id int32) (*entity.Tag, error) {
	tagDAL := dal.GetQueryByCtx(ctx).Tag
	tag, err := tagDAL.WithContext(ctx).Where(tagDAL.ID.Eq(id)).First()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return tag, nil
}

func (t *tagServiceImpl) GetTagByName(ctx context.Context, name string) (*entity.Tag, error) {
	tagDAL := dal.GetQueryByCtx(ctx).Tag
	tag, err := tagDAL.WithContext(ctx).Where(tagDAL.Name.Eq(name)).First()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return tag, nil
}

func (t *tagServiceImpl) GetTagBySlug(ctx context.Context, slug string) (*entity.Tag, error) {
	tagDAL := dal.GetQueryByCtx(ctx).Tag
	tag, err := tagDAL.WithContext(ctx).Where(tagDAL.Slug.Eq(slug)).First()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return tag, nil
}

func (t *tagServiceImpl) GetTagsCount(ctx context.Context) (int64, error) {
	tagDAL := dal.GetQueryByCtx(ctx).Tag
	count, err := tagDAL.WithContext(ctx).Count()
	if err != nil {
		return 0, WrapDBErr(err)
	}
	return count, nil
}

func (t *tagServiceImpl) ConvertToTagDTO(ctx context.Context, tag *entity.Tag) (*dto.Tag, error) {
	tagDTO := &dto.Tag{
		ID:         tag.ID,
		Name:       tag.Name,
		Slug:       tag.Slug,
		Thumbnail:  tag.Thumbnail,
		CreateTime: tag.CreateTime.UnixMilli(),
		Color:      tag.Color,
	}

	fullPath := strings.Builder{}
	fullPath.WriteString("/")

	tagPrefix, err := t.OptionService.GetOrByDefaultWithErr(ctx, property.TagsPrefix, property.TagsPrefix.DefaultValue)
	if err != nil {
		return nil, err
	}

	fullPath.WriteString(tagPrefix.(string))
	fullPath.WriteString("/")
	fullPath.WriteString(tag.Slug)
	tagDTO.FullPath = fullPath.String()

	return tagDTO, nil
}

func (t *tagServiceImpl) ConvertToTagDTOs(ctx context.Context, tags []*entity.Tag) ([]*dto.Tag, error) {
	tagPrefix, err := t.OptionService.GetOrByDefaultWithErr(ctx, property.TagsPrefix, property.TagsPrefix.DefaultValue)
	if err != nil {
		return nil, err
	}

	tagDTOs := make([]*dto.Tag, 0, len(tags))
	for _, tag := range tags {
		fullPath := strings.Builder{}
		fullPath.WriteString("/")
		fullPath.WriteString(tagPrefix.(string))
		fullPath.WriteString("/")
		fullPath.WriteString(tag.Slug)
		tagDTO := &dto.Tag{
			ID:         tag.ID,
			Name:       tag.Name,
			Slug:       tag.Slug,
			Thumbnail:  tag.Thumbnail,
			CreateTime: tag.CreateTime.UnixMilli(),
			FullPath:   fullPath.String(),
			Color:      tag.Color,
		}
		tagDTOs = append(tagDTOs, tagDTO)
	}
	return tagDTOs, nil
}

func (t *tagServiceImpl) ConvertToTagWithPostCountDTO(ctx context.Context, tag *entity.Tag) (*dto.TagWithPostCount, error) {
	tagDTO, err := t.ConvertToTagDTO(ctx, tag)
	if err != nil {
		return nil, err
	}
	postTagDAL := dal.GetQueryByCtx(ctx).PostTag
	count, err := postTagDAL.WithContext(ctx).Where(postTagDAL.TagID.Eq(tagDTO.ID)).Count()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	tagWithPostCountDTO := &dto.TagWithPostCount{
		Tag:       tagDTO,
		PostCount: count,
	}
	return tagWithPostCountDTO, nil
}

func (t *tagServiceImpl) ConvertToTagWithPostCountDTOs(ctx context.Context, tags []*entity.Tag) ([]*dto.TagWithPostCount, error) {
	tagDTOs, err := t.ConvertToTagDTOs(ctx, tags)
	if err != nil {
		return nil, err
	}
	tagWithPostCountDOTs := make([]*dto.TagWithPostCount, 0)
	postTagDAL := dal.GetQueryByCtx(ctx).PostTag
	for _, tagDTO := range tagDTOs {
		count, err := postTagDAL.WithContext(ctx).Where(postTagDAL.TagID.Eq(tagDTO.ID)).Count()
		if err != nil {
			return nil, WrapDBErr(err)
		}
		tagWithPostCountDTO := &dto.TagWithPostCount{
			Tag:       tagDTO,
			PostCount: count,
		}
		tagWithPostCountDOTs = append(tagWithPostCountDOTs, tagWithPostCountDTO)
	}
	return tagWithPostCountDOTs, nil
}
