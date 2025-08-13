package vo

import "dash/model/dto"

type Categories struct {
	*dto.Page
}

type Category struct {
	*dto.Category
	Posts []*Post `json:"posts"`
}
