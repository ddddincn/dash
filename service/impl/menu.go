package impl

import (
	"context"
	"dash/dal"
	"dash/model/dto"
	"dash/model/entity"
	"dash/model/param"
	"dash/service"
	"time"
)

type menuServiceImpl struct {
}

func NewMenuService() service.MenuService {
	return &menuServiceImpl{}
}

func (m *menuServiceImpl) Create(ctx context.Context, menuParam *param.Menu) (*entity.Menu, error) {
	menu := &entity.Menu{
		CreateTime: time.Now(),
		Name:       menuParam.Name,
		URL:        menuParam.URL,
		Icon:       menuParam.Icon,
		Priority:   menuParam.Priority,
		ParentID:   menuParam.ParentID,
		Target:     menuParam.Target,
	}
	menuDAL := dal.GetQueryByCtx(ctx).Menu
	err := menuDAL.WithContext(ctx).Create(menu)
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return menu, nil
}

func (m *menuServiceImpl) List(ctx context.Context, sort *param.Sort) ([]*entity.Menu, error) {
	menuDAL := dal.GetQueryByCtx(ctx).Menu
	menuDO := menuDAL.WithContext(ctx)
	err := BuildSort(sort, &menuDAL, &menuDO)
	if err != nil {
		return nil, err
	}
	menus, err := menuDO.Find()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	return menus, nil
}

func (m *menuServiceImpl) ConvertToMenuDTO(ctx context.Context, menu *entity.Menu) (*dto.Menu, error) {
	return &dto.Menu{
		ID:       menu.ID,
		Name:     menu.Name,
		URL:      menu.URL,
		Priority: menu.Priority,
		Target:   menu.Target,
		Icon:     menu.Icon,
		ParentID: menu.ParentID,
		Team:     menu.Team,
	}, nil
}

func (m *menuServiceImpl) ConvertToMenuDTOs(ctx context.Context, menus []*entity.Menu) ([]*dto.Menu, error) {
	menuDTOs := make([]*dto.Menu, 0)

	for _, link := range menus {
		menuDTO, err := m.ConvertToMenuDTO(ctx, link)
		if err != nil {
			return nil, err
		}

		menuDTOs = append(menuDTOs, menuDTO)
	}
	return menuDTOs, nil
}
