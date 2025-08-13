package impl

import (
	"context"
	"dash/consts"
	"dash/dal"
	"dash/model/entity"
	"dash/model/param"
	"dash/model/property"
	"dash/service"
	"dash/utils/xerr"
	"database/sql/driver"
	"errors"

	"gorm.io/gen/field"
	"gorm.io/gorm"
)

type postServiceImpl struct {
	service.BasePostService
	service.OptionService
}

func NewPostService(
	basePostService service.BasePostService,
	optionService service.OptionService,
) service.PostService {
	return &postServiceImpl{
		BasePostService: basePostService,
		OptionService:   optionService,
	}
}

func (p *postServiceImpl) Page(ctx context.Context, postQuery param.PostQuery) ([]*entity.Post, int64, error) {
	if postQuery.PageNum < 0 || postQuery.PageSize < 0 { // 验证分页参数有效性
		return nil, 0, xerr.BadParam.New("").WithStatus(xerr.StatusBadRequest).WithMsg("Paging parameter error")
	}

	// 初始化数据访问层
	postDAL := dal.GetQueryByCtx(ctx).Post                                         // 获取文章数据访问层对象
	postCategoryDAL := dal.GetQueryByCtx(ctx).PostCategory                         // 获取文章分类数据访问层对象
	postTagDAL := dal.GetQueryByCtx(ctx).PostTag                                   // 获取文章标签数据访问层对象
	postDo := postDAL.WithContext(ctx).Where(postDAL.Type.Eq(consts.PostTypePost)) //设置基础查询条件，只查询类型为PostTypePost的记录
	err := BuildSort(postQuery.Sort, &postDAL, &postDo)
	if err != nil {
		return nil, 0, err
	}

	if postQuery.Keyword != nil { //根据关键词进行模糊查询
		keyword := "%" + *postQuery.Keyword + "%"
		postDo = postDo.Where(field.Or(postDAL.Title.Like(keyword), postDAL.OriginalContent.Like(keyword)))
	}
	if len(postQuery.Statuses) > 0 { // 文章状态过滤，只查询指定状态的文章
		statuesValue := make([]driver.Valuer, len(postQuery.Statuses)) // 初始化状态值切片
		for i, status := range postQuery.Statuses {
			statuesValue[i] = driver.Valuer(status)
		}
		postDo = postDo.Where(postDAL.Status.In(statuesValue...))
	}
	if postQuery.CategoryID != nil { // 文章分类过滤，只查询指定分类的文章
		postDo.Join(&entity.PostCategory{}, postDAL.ID.EqCol(postCategoryDAL.PostID)).Where(postCategoryDAL.CategoryID.Eq(*postQuery.CategoryID))
	}
	if postQuery.TagID != nil { // 文章标签过滤，只查询指定标签的文章
		postDo.Join(&entity.PostTag{}, postDAL.ID.EqCol(postTagDAL.PostID)).Where(postTagDAL.TagID.Eq(*postQuery.TagID))
	}

	posts, totalCount, err := postDo.FindByPage(postQuery.PageNum*postQuery.PageSize, postQuery.PageSize) // 分页查询文章列表
	if err != nil {
		return nil, 0, WrapDBErr(err)
	}
	return posts, totalCount, nil
}

func (p *postServiceImpl) GetPrevPosts(ctx context.Context, post *entity.Post, size int) ([]*entity.Post, error) {
	postSort := p.OptionService.GetOrByDefault(ctx, property.IndexSort)
	postDAL := dal.GetQueryByCtx(ctx).Post
	postDO := postDAL.WithContext(ctx).Where(postDAL.Status.Eq(consts.PostStatusPublished), postDAL.Type.Eq(consts.PostTypePost))

	switch postSort {
	case "create_time":
		postDO = postDO.Where(postDAL.CreateTime.Gt(post.CreateTime)).Order(postDAL.CreateTime)
	case "edit_time":
		editTime := post.CreateTime
		if post.EditTime != nil {
			editTime = *post.EditTime
		}
		postDO = postDO.Where(postDAL.EditTime.Gt(editTime)).Order(postDAL.EditTime)
	case "visits":
		postDO = postDO.Where(postDAL.Visits.Gt(post.Visits)).Order(postDAL.EditTime)
	default:
		return nil, nil
	}

	posts, err := postDO.Find()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, WrapDBErr(err)
	}
	return posts, nil
}

func (p *postServiceImpl) GetNextPosts(ctx context.Context, post *entity.Post, size int) ([]*entity.Post, error) {
	postSort := p.OptionService.GetOrByDefault(ctx, property.IndexSort)
	postDAL := dal.GetQueryByCtx(ctx).Post
	postDO := postDAL.WithContext(ctx).Where(postDAL.Status.Eq(consts.PostStatusPublished), postDAL.Type.Eq(consts.PostTypePost))

	switch postSort {
	case "create_time":
		postDO = postDO.Where(postDAL.CreateTime.Lt(post.CreateTime)).Order(postDAL.CreateTime.Desc())
	case "edit_time":
		editTime := post.CreateTime
		if post.EditTime != nil {
			editTime = *post.EditTime
		}
		postDO = postDO.Where(postDAL.EditTime.Lt(editTime)).Order(postDAL.EditTime.Desc())
	case "visits":
		postDO = postDO.Where(postDAL.Visits.Lt(post.Visits)).Order(postDAL.EditTime.Desc())
	default:
		return nil, nil
	}

	posts, err := postDO.Find()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return posts, nil
}

func (p *postServiceImpl) GetPostCountByStatus(ctx context.Context, status consts.PostStatus) (int64, error) {
	postDAL := dal.GetQueryByCtx(ctx).Post
	count, err := postDAL.WithContext(ctx).Where(postDAL.Type.Eq(consts.PostTypePost), postDAL.Status.Eq(status)).Count()
	if err != nil {
		return 0, WrapDBErr(err)
	}
	return count, nil
}

func (p *postServiceImpl) GetVisitCount(ctx context.Context) (int64, error) {
	var count float64
	postDAL := dal.GetQueryByCtx(ctx).Post
	err := postDAL.WithContext(ctx).Select(postDAL.Visits.Sum().IfNull(0)).Where(postDAL.Type.Eq(consts.PostTypePost), postDAL.Status.Eq(consts.PostStatusPublished)).Scan(&count)
	if err != nil {
		return 0, WrapDBErr(err)
	}
	return int64(count), nil
}

func (p *postServiceImpl) GetLikeCount(ctx context.Context) (int64, error) {
	var count float64
	postDAL := dal.GetQueryByCtx(ctx).Post
	err := postDAL.WithContext(ctx).Select(postDAL.Likes.Sum().IfNull(0)).Where(postDAL.Type.Eq(consts.PostTypePost), postDAL.Status.Eq(consts.PostStatusPublished)).Scan(&count)
	if err != nil {
		return 0, WrapDBErr(err)
	}
	return int64(count), nil
}
