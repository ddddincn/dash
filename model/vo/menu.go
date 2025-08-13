package vo

import "dash/model/dto"

type Menu struct {
	dto.Menu
	Children []*Menu `json:"children"`
}
