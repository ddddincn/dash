package impl

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"dash/consts"
	"dash/dal"
	"dash/model/entity"
	"dash/model/param"
	"dash/model/property"
	"dash/service"
	"dash/utils"
	"dash/utils/xerr"
)

type installServiceImpl struct {
	OptionService   service.OptionService
	UerService      service.UserService
	CategoryService service.CategoryService
	PostService     service.PostService
	// SheetService    service.SheetService
	MenuService service.MenuService
}

func NewInstallService(
	optionService service.OptionService,
	uerService service.UserService,
	categoryService service.CategoryService,
	postService service.PostService,
	menuService service.MenuService,
) service.InstallService {
	return &installServiceImpl{
		OptionService:   optionService,
		UerService:      uerService,
		CategoryService: categoryService,
		PostService:     postService,
		MenuService:     menuService,
	}
}

func (i *installServiceImpl) InstallBlog(ctx context.Context, installParam *param.Install) error {
	isInstalled, err := i.OptionService.GetOrByDefaultWithErr(ctx, property.IsInstalled, false)
	if err != nil {
		return nil
	}
	if isInstalled.(bool) {
		return xerr.BadParam.New("").WithStatus(xerr.StatusBadRequest).WithMsg("Blog has been installed")
	}
	// var user *entity.User
	err = dal.Transaction(ctx, func(txCtx context.Context) error {
		if err := i.createJWTSecret(txCtx); err != nil {
			return err
		}
		if err := i.createDefaultSetting(txCtx, installParam); err != nil {
			return err
		}
		_, err = i.createUser(txCtx, &installParam.User)
		if err != nil {
			return err
		}
		category, err := i.createDefaultCategory(txCtx)
		if err != nil {
			return err
		}
		_, err = i.createDefaultPost(txCtx, category)
		if err != nil {
			return err
		}
		_, err = i.createDefaultSheet(txCtx)
		if err != nil {
			return err
		}
		err = i.createDefaultMenu(txCtx)
		return err
	})
	if err != nil {
		return err
	}
	return err
}

func (i *installServiceImpl) createDefaultSetting(ctx context.Context, installParam *param.Install) error {
	optionMap := make(map[string]string)
	optionMap[property.IsInstalled.KeyValue] = "true"
	optionMap[property.BlogTitle.KeyValue] = installParam.Title
	if installParam.URL == "" {
		blogURL, err := i.OptionService.GetBlogBaseURL(ctx)
		if err != nil {
			return err
		}
		optionMap[property.BlogURL.KeyValue] = blogURL
	} else {
		optionMap[property.BlogURL.KeyValue] = installParam.URL
	}
	optionMap[property.BirthDay.KeyValue] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	err := i.OptionService.Save(ctx, optionMap)
	return err
}

func (i *installServiceImpl) createUser(ctx context.Context, user *param.User) (*entity.User, error) {
	emailMd5 := md5.Sum([]byte(user.Email))
	avatar := "//cn.gravatar.com/avatar/" + hex.EncodeToString(emailMd5[:]) + "?s=256&d=mm"
	user.Avatar = avatar
	userEntity, err := i.UerService.Create(ctx, user)
	return userEntity, err
}

func (i *installServiceImpl) createDefaultCategory(ctx context.Context) (*entity.Category, error) {
	categoryDal := dal.GetQueryByCtx(ctx).Category
	count, err := categoryDal.WithContext(ctx).Count()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	if count > 0 {
		return nil, nil
	}
	categoryParam := param.Category{
		Name:        "默认分类",
		Slug:        "default",
		Description: "这是你的默认分类，如不需要，删除即可",
	}
	category, err := i.CategoryService.Create(ctx, &categoryParam)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (i *installServiceImpl) createDefaultPost(ctx context.Context, category *entity.Category) (*entity.Post, error) {
	if category == nil {
		return nil, nil
	}
	postDAL := dal.GetQueryByCtx(ctx).Post
	count, err := postDAL.WithContext(ctx).Where(postDAL.Status.Eq(consts.PostStatusPublished)).Count()
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, nil
	}
	content := `
## Hello Dash

如果你看到了这一篇文章，那么证明你已经安装成功了，希望能够使用愉快。

> 这是一篇自动生成的文章，请删除这篇文章之后开始你的创作吧！
`
	formatContent := `<h2>Hello Dash</h2>
	<p>如果你看到了这一篇文章，那么证明你已经安装成功了，希望能够使用愉快。</p>
	
	<blockquote>
	<p>这是一篇自动生成的文章，请删除这篇文章之后开始你的创作吧！</p>
	</blockquote>
	`
	postParam := param.Post{
		Title:           "Hello Dash",
		Status:          consts.PostStatusPublished,
		Slug:            "hello-dash",
		OriginalContent: content,
		Content:         formatContent,
		CategoryIDs:     []int32{category.ID},
	}
	return i.PostService.Create(ctx, &postParam, consts.PostTypePost)
}

func (i *installServiceImpl) createDefaultSheet(ctx context.Context) (*entity.Post, error) {
	postDAL := dal.GetQueryByCtx(ctx).Post
	count, err := postDAL.WithContext(ctx).Where(postDAL.Status.Eq(consts.PostStatusPublished), postDAL.Type.Eq(consts.PostTypeSheet)).Count()
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, nil
	}
	originalContent := "## 关于页面 \n\n" +
		" 这是一个自定义页面，你可以在后台的 `页面` -> `所有页面` -> `自定义页面` 找到它，" +
		"你可以用于新建关于页面、留言板页面等等。发挥你自己的想象力！\n\n" +
		"> 这是一篇自动生成的页面，你可以在后台删除它。"
	formatContent := `<h2>关于页面</h2>
<p>这是一个自定义页面，你可以在后台的 <code>页面</code> -&gt; <code>所有页面</code> -&gt; <code>自定义页面</code> 找到它，你可以用于新建关于页面、留言板页面等等。发挥你自己的想象力！</p>
<blockquote>
<p>这是一篇自动生成的页面，你可以在后台删除它。</p>
</blockquote>`
	sheetParam := param.Post{
		Title:           "关于页面",
		Status:          consts.PostStatusPublished,
		Slug:            "about",
		OriginalContent: originalContent,
		Content:         formatContent,
	}
	return i.PostService.Create(ctx, &sheetParam, consts.PostTypeSheet)
}

func (i *installServiceImpl) createDefaultMenu(ctx context.Context) error {
	menuIndex := &param.Menu{
		Name:     "首页",
		URL:      "/",
		Priority: 1,
	}
	menuArchive := &param.Menu{
		Name:     "文章归档",
		URL:      "/archives",
		Priority: 2,
	}
	menuCategory := &param.Menu{
		Name:     "分类目录",
		URL:      "/categories",
		Priority: 3,
	}
	menuSheet := &param.Menu{
		Name:     "关于页面",
		URL:      "/s/about",
		Priority: 4,
	}
	createMenu := func(menu *param.Menu, err error) error {
		if err != nil {
			return err
		}
		_, err = i.MenuService.Create(ctx, menu)
		return err
	}
	err := createMenu(menuIndex, nil)
	err = createMenu(menuArchive, err)
	err = createMenu(menuCategory, err)
	err = createMenu(menuSheet, err)
	return err
}

func (i *installServiceImpl) createJWTSecret(ctx context.Context) error {
	access_secret := &strings.Builder{}
	refresh_secret := &strings.Builder{}
	access_secret.Grow(256)
	refresh_secret.Grow(256)
	for i := 0; i < 8; i++ {
		access_secret.WriteString(utils.GenUUIDWithOutDash())
		refresh_secret.WriteString(utils.GenUUIDWithOutDash())
	}
	m := map[string]string{
		property.JWTAccessSecret.KeyValue:  access_secret.String(),
		property.JWTRefreshSecret.KeyValue: refresh_secret.String(),
	}
	return i.OptionService.Save(ctx, m)
}
