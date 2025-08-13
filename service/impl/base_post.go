package impl

import (
	"context"
	"dash/consts"
	"dash/dal"
	"dash/model/entity"
	"dash/model/param"
	"dash/service"
	"dash/utils"
	"dash/utils/xerr"
	"regexp"
	"time"
)

type basePostServiceImpl struct {
	OptionService service.OptionService
}

func NewBasePostService(
	optionService service.OptionService,
) service.BasePostService {
	return &basePostServiceImpl{
		OptionService: optionService,
	}
}

func (b *basePostServiceImpl) Create(ctx context.Context, postParam *param.Post, postType consts.PostType) (*entity.Post, error) {
	post, err := b.ConvertToEntity(ctx, postParam, postType)
	if err != nil {
		return nil, err
	}
	categoryIDs := postParam.CategoryIDs
	tagIDs := postParam.TagIDs
	err = dal.Transaction(ctx, func(txCtx context.Context) error {
		now := time.Now()
		post.CreateTime = now

		query := dal.GetQueryByCtx(txCtx)
		postDAL := query.Post
		categoryDAL := query.Category
		tagDAL := query.Tag
		postCategoryDAL := query.PostCategory
		postTagDAL := query.PostTag

		slugCount, err := postDAL.WithContext(txCtx).Where(postDAL.Slug.Eq(post.Slug)).Count()
		if err != nil {
			return WrapDBErr(err)
		}
		if slugCount > 0 {
			return xerr.BadParam.New("").WithMsg("post slug already exists").WithStatus(xerr.StatusBadRequest)
		}

		if post.Summary == "" {
			post.Summary = b.generateSummary(ctx, post.FormatContent)
		}

		status := post.Status
		err = postDAL.WithContext(txCtx).Create(post)
		if err != nil {
			return WrapDBErr(err)
		}
		// gorm not insert zero value: https://gorm.io/docs/create.html
		if status == consts.PostStatusPublished {
			_, err = postDAL.WithContext(txCtx).Where(postDAL.ID.Eq(post.ID)).UpdateColumnSimple(postDAL.Status.Value(status))
			if err != nil {
				return WrapDBErr(err)
			}
			post.Status = status
		}

		if len(categoryIDs) > 0 {
			categoryCount, err := categoryDAL.WithContext(txCtx).Where(categoryDAL.ID.In(categoryIDs...)).Count()
			if err != nil {
				return WrapDBErr(err)
			}
			if int(categoryCount) != len(categoryIDs) {
				return xerr.BadParam.New("").WithMsg("category not exist").WithStatus(xerr.StatusBadRequest)
			}
			pcs := make([]*entity.PostCategory, 0, len(categoryIDs))
			for _, categoryID := range categoryIDs {
				pc := &entity.PostCategory{
					CreateTime: now,
					CategoryID: categoryID,
					PostID:     post.ID,
				}
				pcs = append(pcs, pc)
			}
			err = postCategoryDAL.WithContext(txCtx).Create(pcs...)
			if err != nil {
				return WrapDBErr(err)
			}

		}

		if len(tagIDs) > 0 {
			tagCount, err := tagDAL.WithContext(txCtx).Where(tagDAL.ID.In(tagIDs...)).Count()
			if err != nil {
				return WrapDBErr(err)
			}
			if int(tagCount) != len(tagIDs) {
				return xerr.BadParam.New("").WithMsg("tag not exist").WithStatus(xerr.StatusBadRequest)
			}
			pts := make([]*entity.PostTag, 0, len(tagIDs))
			for _, tagID := range tagIDs {
				pts = append(pts, &entity.PostTag{
					CreateTime: now,
					PostID:     post.ID,
					TagID:      tagID,
				})
			}
			err = postTagDAL.WithContext(txCtx).Create(pts...)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (b *basePostServiceImpl) DeleteByID(ctx context.Context, id int32) error {
	err := dal.Transaction(ctx, func(txCtx context.Context) error {
		query := dal.GetQueryByCtx(txCtx)
		postDAL := query.Post
		postCategoryDAL := query.PostCategory
		postTagDAL := query.PostTag

		// delete post
		deleteResult, err := postDAL.WithContext(ctx).Where(postDAL.ID.Eq(id)).Delete()
		if err != nil {
			return WrapDBErr(err)
		}
		if deleteResult.RowsAffected != 1 { // deleted post count must = 1
			return xerr.NoType.New("").WithMsg("delete post failed")
		}
		_, err = postTagDAL.WithContext(ctx).Where(postTagDAL.PostID.Eq(id)).Delete()
		if err != nil {
			return WrapDBErr(err)
		}
		_, err = postCategoryDAL.WithContext(ctx).Where(postCategoryDAL.PostID.Eq(id)).Delete()
		if err != nil {
			return WrapDBErr(err)
		}
		return nil
	})
	return err
}

func (b *basePostServiceImpl) DeleteBatchByID(ctx context.Context, ids []int32) error {
	err := dal.Transaction(ctx, func(txCtx context.Context) error {
		for _, id := range ids {
			err := b.DeleteByID(ctx, id)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (b *basePostServiceImpl) UpdateByID(ctx context.Context, id int32, postParam *param.Post, postType consts.PostType) (*entity.Post, error) {
	post, err := b.ConvertToEntity(ctx, postParam, postType)
	if err != nil {
		return nil, err
	}
	categoryIDs := postParam.CategoryIDs
	tagIDs := postParam.TagIDs
	err = dal.Transaction(ctx, func(txCtx context.Context) error {
		now := time.Now()
		post.UpdateTime = &now

		query := dal.GetQueryByCtx(txCtx)
		postDAL := query.Post
		categoryDAL := query.Category
		tagDAL := query.Tag
		postCategoryDAL := query.PostCategory
		postTagDAL := query.PostTag

		// determine if the post ID exists
		_, err := postDAL.WithContext(txCtx).Where(postDAL.ID.Eq(id)).First()
		if err != nil {
			return WrapDBErr(err)
		}

		// determine if the post slug exists
		slugCount, err := postDAL.WithContext(txCtx).
			Where(
				postDAL.Slug.Neq(post.Slug),
				postDAL.Slug.Eq(post.Slug),
			).Count()
		if err != nil {
			return WrapDBErr(err)
		}
		// if exists, determine is itself?
		if slugCount != 0 {
			return xerr.BadParam.New("invalid parameter").WithMsg("post slug already exists").WithStatus(xerr.StatusBadRequest)
		}

		// generateSummary
		if post.Summary == "" {
			post.Summary = b.generateSummary(ctx, post.FormatContent)
		}

		// update post
		updateResult, err := postDAL.WithContext(txCtx).Where(postDAL.ID.Eq(id)).Updates(post)
		if err != nil {
			return WrapDBErr(err)
		}
		if updateResult.RowsAffected != 1 {
			return xerr.NoType.New("").WithMsg("update post failed")
		}

		_, err = postCategoryDAL.WithContext(txCtx).Where(postCategoryDAL.PostID.Eq(id)).Delete()
		if err != nil {
			return WrapDBErr(err)
		}

		// delete original post-tag info
		_, err = postTagDAL.WithContext(txCtx).Where(postTagDAL.PostID.Eq(id)).Delete()
		if err != nil {
			return WrapDBErr(err)
		}

		// re-record post-category info
		if len(categoryIDs) > 0 {
			categoryCount, err := categoryDAL.WithContext(txCtx).Where(categoryDAL.ID.In(categoryIDs...)).Count()
			if err != nil {
				return WrapDBErr(err)
			}
			if int(categoryCount) != len(categoryIDs) {
				return xerr.BadParam.New("").WithMsg("category not exist").WithStatus(xerr.StatusBadRequest)
			}
			pcs := make([]*entity.PostCategory, 0, len(categoryIDs))
			for _, categoryID := range categoryIDs {
				pc := &entity.PostCategory{
					CreateTime: now,
					PostID:     id,
					CategoryID: categoryID,
				}
				pcs = append(pcs, pc)
			}
			err = postCategoryDAL.WithContext(txCtx).Create(pcs...)
			if err != nil {
				return WrapDBErr(err)
			}

		}
		// re-record post-info info
		if len(tagIDs) > 0 {
			tagCount, err := tagDAL.WithContext(txCtx).Where(tagDAL.ID.In(tagIDs...)).Count()
			if err != nil {
				return WrapDBErr(err)
			}
			if int(tagCount) != len(tagIDs) {
				return xerr.BadParam.New("").WithMsg("tag not exist").WithStatus(xerr.StatusBadRequest)
			}

			pts := make([]*entity.PostTag, 0, len(tagIDs))
			for _, tagID := range tagIDs {
				pts = append(pts, &entity.PostTag{
					CreateTime: now,
					PostID:     id,
					TagID:      tagID,
				})
			}
			err = postTagDAL.WithContext(txCtx).Create(pts...)
			if err != nil {
				return err
			}
		}
		post, err = postDAL.WithContext(ctx).Where(postDAL.ID.Eq(id)).First()
		if err != nil {
			return WrapDBErr(err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (b *basePostServiceImpl) UpdateStatusByID(ctx context.Context, id int32, status consts.PostStatus) (*entity.Post, error) {
	if id < 0 || status < consts.PostStatusPublished || status > consts.PostStatusIntimate {
		return nil, xerr.BadParam.New("invalid parameter").WithMsg("post ID or status parameter error").WithStatus(xerr.StatusBadRequest)
	}

	postDAL := dal.GetQueryByCtx(ctx).Post
	post, err := postDAL.WithContext(ctx).Where(postDAL.ID.Eq(id)).First()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	updateResult, err := postDAL.WithContext(ctx).Where(postDAL.ID.Eq(id)).UpdateColumnSimple(postDAL.Status.Value(status))
	if err != nil {
		return nil, WrapDBErr(err)
	}
	if updateResult.RowsAffected != 1 {
		return nil, xerr.NoType.New("update post status failed ID=%v", id).WithMsg("update post status failed")
	}
	post.Status = status
	return post, nil
}

func (b *basePostServiceImpl) UpdateStatusBatch(ctx context.Context, ids []int32, status consts.PostStatus) ([]*entity.Post, error) {
	if status < consts.PostStatusPublished || status > consts.PostStatusIntimate {
		return nil, xerr.BadParam.New("").WithMsg("postID or status parameter error").WithStatus(xerr.StatusBadRequest)
	}

	uniquePostIDMap := make(map[int32]struct{})
	for _, postID := range ids {
		uniquePostIDMap[postID] = struct{}{}
	}
	uniqueIDs := make([]int32, 0)
	for postID := range uniquePostIDMap {
		uniqueIDs = append(uniqueIDs, postID)
	}
	err := dal.GetQueryByCtx(ctx).Transaction(func(tx *dal.Query) error {
		postDAL := tx.Post
		updateResult, err := postDAL.WithContext(ctx).Where(postDAL.ID.In(uniqueIDs...)).UpdateColumnSimple(postDAL.Status.Value(status))
		if err != nil {
			return WrapDBErr(err)
		}
		if updateResult.RowsAffected != int64(len(uniqueIDs)) {
			return xerr.NoType.New("").WithMsg("update post status failed")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	postDAL := dal.GetQueryByCtx(ctx).Post
	posts, err := postDAL.WithContext(ctx).Where(postDAL.ID.In(uniqueIDs...)).Find()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return posts, nil
}

func (b *basePostServiceImpl) List(ctx context.Context, sort *param.Sort) ([]*entity.Post, error) {
	postDAL := dal.GetQueryByCtx(ctx).Post
	postDO := postDAL.WithContext(ctx)
	err := BuildSort(sort, &postDAL, &postDO)
	if err != nil {
		return nil, err
	}
	posts, err := postDAL.WithContext(ctx).Where().Find()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return posts, nil
}

func (b *basePostServiceImpl) ListByIDs(ctx context.Context, ids []int32) ([]*entity.Post, error) {
	if len(ids) == 0 {
		return make([]*entity.Post, 0), nil
	}
	postDAL := dal.GetQueryByCtx(ctx).Post
	posts, err := postDAL.WithContext(ctx).Where(postDAL.ID.In(ids...)).Find()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return posts, nil
}

func (b *basePostServiceImpl) GetPostByID(ctx context.Context, id int32) (*entity.Post, error) {
	postDAL := dal.GetQueryByCtx(ctx).Post
	post, err := postDAL.WithContext(ctx).Where(postDAL.ID.Eq(id)).First()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return post, nil
}

func (b *basePostServiceImpl) GetPostBySlug(ctx context.Context, slug string) (*entity.Post, error) {
	postDAL := dal.GetQueryByCtx(ctx).Post
	post, err := postDAL.WithContext(ctx).Where(postDAL.Slug.Eq(slug)).First()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return post, nil
}

func (b *basePostServiceImpl) GetPostsCount(ctx context.Context) (int64, error) {
	postDAL := dal.GetQueryByCtx(ctx).Post
	count, err := postDAL.WithContext(ctx).Count()
	if err != nil {
		return 0, WrapDBErr(err)
	}
	return count, nil
}

func (b *basePostServiceImpl) ConvertToEntity(ctx context.Context, postParam *param.Post, postType consts.PostType) (*entity.Post, error) {
	post := &entity.Post{
		Type:            postType,
		OriginalContent: postParam.OriginalContent,
		Thumbnail:       postParam.Thumbnail,
		Title:           postParam.Title,
		TopPriority:     postParam.TopPriority,
		Status:          postParam.Status,
		Summary:         postParam.Summary,
		FormatContent:   postParam.Content,
	}
	if postParam.EditorType != nil {
		post.EditorType = *postParam.EditorType
	} else {
		post.EditorType = consts.EditorTypeMarkdown
	}
	post.WordCount = utils.HTMLFormatWordCount(post.FormatContent)
	if postParam.Slug == "" {
		post.Slug = utils.Slug(postParam.Title)
	} else {
		post.Slug = utils.Slug(postParam.Slug)
	}

	return post, nil
}

var summaryPattern = regexp.MustCompile(`[\t\r\n]`)

func (b *basePostServiceImpl) generateSummary(ctx context.Context, htmlContent string) string {
	text := utils.CleanHTMLTag(htmlContent)
	text = summaryPattern.ReplaceAllString(text, "")
	summaryLength := b.OptionService.GetPostSummaryLength(ctx)
	end := summaryLength
	textRune := []rune(text)
	if len(textRune) < end {
		end = len(textRune)
	}
	return string(textRune[:end])
}
