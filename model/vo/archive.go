package vo

import "dash/model/dto"

type Archives struct {
	*dto.Page
}

type Archive struct {
	Year  int     `json:"year"`
	Posts []*Post `json:"posts"`
}
