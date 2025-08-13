package service

import (
	"context"
	"dash/model/dto"
	"dash/model/entity"
	"dash/model/param"
)

type MenuService interface {
	Create(ctx context.Context, menuParam *param.Menu) (*entity.Menu, error)
	List(ctx context.Context, sort *param.Sort) ([]*entity.Menu, error)
	ConvertToMenuDTO(ctx context.Context, menu *entity.Menu) (*dto.Menu, error)
	ConvertToMenuDTOs(ctx context.Context, menus []*entity.Menu) ([]*dto.Menu, error)
}
