package service

import (
	"context"
	"dash/model/entity"
)

type ThemeService interface {
	ListByThemeID(ctx context.Context, themeID string) ([]*entity.ThemeSetting, error)
	GetThemeSettingMapByThemeID(ctx context.Context, themeID string) (map[string]interface{}, error)
	// ConvertToThemeSettingDTO(ctx context.Context, themeSetting *entity.ThemeSetting) (*dto.ThemeSetting, error)
	// ConvertToThemeSettingDTOs(ctx context.Context, themeSettings []*entity.ThemeSetting) ([]*dto.ThemeSetting, error)
}
