package handler

import (
	"dash/model/property"
	"dash/model/vo"
	"dash/service"
	"dash/utils"

	"github.com/gin-gonic/gin"
)

type ThemeHandler struct {
	OptionService service.OptionService
	UserService   service.UserService
	ThemeService  service.ThemeService
	// MenuService    service.MenuService
	// PostTagService service.PostTagService
}

func NewThemeHandler(optionService service.OptionService, userService service.UserService, themeService service.ThemeService) *ThemeHandler {
	return &ThemeHandler{
		OptionService: optionService,
		UserService:   userService,
		ThemeService:  themeService,
		// MenuService:    menuService,
		// PostTagService: postTagService,
	}
}

func (l *ThemeHandler) GetThemeSettings(ctx *gin.Context) (interface{}, error) {
	blogURL, err := l.OptionService.GetBlogBaseURL(ctx)
	if err != nil {
		return nil, err
	}
	blogTitle := l.OptionService.GetOrByDefault(ctx, property.BlogTitle)
	user, err := l.UserService.GetFirst(ctx)
	if err != nil {
		return nil, err
	}
	userDTO := l.UserService.ConvertToUserDTO(user)

	themeID, err := utils.ParamString(ctx, "themeID")
	if err != nil {
		return nil, err
	}

	themeSettingsMap, err := l.ThemeService.GetThemeSettingMapByThemeID(ctx, themeID)
	if err != nil {
		return nil, err
	}

	// menus, err := l.MenuService.List(ctx, &param.Sort{Fields: []string{"priority"}})
	// if err != nil {
	// 	return nil, err
	// }
	// menuDTOs := l.MenuService.ConvertToMenuDTOs(ctx, menus)

	// tagWithPostCountDTOs, err := l.PostTagService.ListTagWithPostCount(ctx, &param.Sort{Fields: []string{"id"}})
	// if err != nil {
	// 	return nil, err
	// }

	userVO := &vo.User{
		Nickname:    userDTO.Nickname,
		Avatar:      userDTO.Avatar,
		Description: userDTO.Description,
	}

	// 安全地从主题设置映射中获取值，提供默认值以避免空指针异常
	// Safely get values from theme settings map with default values to avoid nil pointer exceptions
	getStringValue := func(key string, defaultValue string) string {
		if value, ok := themeSettingsMap[key]; ok && value != nil {
			if str, ok := value.(string); ok {
				return str
			}
		}
		return defaultValue
	}

	getBoolValue := func(key string, defaultValue bool) bool {
		if value, ok := themeSettingsMap[key]; ok && value != nil {
			if b, ok := value.(bool); ok {
				return b
			}
		}
		return defaultValue
	}

	settingsVO := &vo.Settings{
		Icon:         getStringValue("icon", ""),
		AvatarCircle: getBoolValue("avatar_circle", false),
		SidebarWidth: getStringValue("sidebar_width", "20%"),
		RSS:          getStringValue("rss", ""),
		Twitter:      getStringValue("twitter", ""),
		Facebook:     getStringValue("facebook", ""),
		Instagram:    getStringValue("instagram", ""),
		Weibo:        getStringValue("weibo", ""),
		QQ:           getStringValue("qq", ""),
		Telegram:     getStringValue("telegram", ""),
		Email:        getStringValue("email", ""),
		Github:       getStringValue("github", ""),
	}

	sidebarInfo := &vo.SidebarInfo{
		BlogURL:   blogURL,
		BlogTitle: blogTitle.(string),
		User:      userVO,
		Settings:  settingsVO,
		// Menus:         menuDTOs,
		// Tags:          tagWithPostCountDTOs,
	}
	return sidebarInfo, nil
}
