package impl

import (
	"context"
	"dash/cache"
	"dash/dal"
	"dash/model/entity"
	"dash/service"
	"encoding/json"
)

type themeServiceImpl struct {
}

func NewThemeService() service.ThemeService {
	return &themeServiceImpl{}
}

func (t *themeServiceImpl) ListByThemeID(ctx context.Context, themeID string) ([]*entity.ThemeSetting, error) {
	if themeID == "" {
		return make([]*entity.ThemeSetting, 0), nil
	}
	themeSettings, err := t.getFromCacheMissFromDB(ctx, themeID)
	if err != nil {
		return nil, err
	}
	return themeSettings, nil
}

func (t *themeServiceImpl) GetThemeSettingMapByThemeID(ctx context.Context, themeID string) (map[string]interface{}, error) {
	themeSettings, err := t.getFromCacheMissFromDB(ctx, themeID)
	if err != nil {
		return nil, err
	}
	result := make(map[string]interface{})
	for _, themeSetting := range themeSettings {
		var value interface{}
		if themeSetting.SettingValue == "true" || themeSetting.SettingValue == "false" {
			value = themeSetting.SettingValue == "true"
		} else {
			value = themeSetting.SettingValue
		}
		result[themeSetting.SettingKey] = value
	}
	return result, nil
}

// func (t *themeServiceImpl) ConvertToThemeSettingDTO(ctx context.Context, themeSetting *entity.ThemeSetting) (*dto.ThemeSetting, error) {
// 	if themeSetting == nil {
// 		return nil, nil
// 	}
// 	var value interface{}
// 	if themeSetting.SettingValue == "true" || themeSetting.SettingValue == "false" {
// 		value = themeSetting.SettingValue == "true"
// 	} else {
// 		value = themeSetting.SettingValue
// 	}
// 	return &dto.ThemeSetting{
// 		Key:   themeSetting.SettingKey,
// 		Value: value,
// 	}, nil
// }

// func (t *themeServiceImpl) ConvertToThemeSettingDTOs(ctx context.Context, themeSettings []*entity.ThemeSetting) ([]*dto.ThemeSetting, error) {
// 	result := make([]*dto.ThemeSetting, len(themeSettings))
// 	for i, themeSetting := range themeSettings {
// 		themeSettingDTO, err := t.ConvertToThemeSettingDTO(ctx, themeSetting)
// 		if err != nil {
// 			return nil, err
// 		}
// 		result[i] = themeSettingDTO
// 	}
// 	return result, nil
// }

func (t *themeServiceImpl) getFromCacheMissFromDB(ctx context.Context, themeID string) ([]*entity.ThemeSetting, error) {
	value, ok, err := cache.Get(themeID)
	if err != nil {
		return nil, err
	}
	if ok {
		themeSettings, _err := convertInterfaceToThemeSettings(value)
		if _err != nil {
			return nil, _err
		}
		return themeSettings, nil
	}
	themeSettingsDAL := dal.GetQueryByCtx(ctx).ThemeSetting
	themeSettings, err := themeSettingsDAL.WithContext(ctx).Where(
		themeSettingsDAL.ThemeID.Eq(themeID)).Find()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	cache.SetDefault(themeID, themeSettings)
	return themeSettings, nil
}

func convertInterfaceToThemeSettings(value interface{}) ([]*entity.ThemeSetting, error) {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var result []*entity.ThemeSetting
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, err
	}
	return result, nil
}
