package handler

import (
	"dash/service"

	"github.com/gin-gonic/gin"
)

type MenuHandler struct {
	MenuService service.MenuService
}

func NewMenuHandler(menuService service.MenuService) *MenuHandler {
	return &MenuHandler{
		MenuService: menuService,
	}
}

func (m *MenuHandler) ListMenus(ctx *gin.Context) (interface{}, error) {
	menus, err := m.MenuService.List(ctx, nil)
	if err != nil {
		return nil, err
	}
	menusDTOs, err := m.MenuService.ConvertToMenuDTOs(ctx, menus)
	if err != nil {
		return nil, err
	}
	return menusDTOs, nil
}
