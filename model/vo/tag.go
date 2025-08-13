package vo

import "dash/model/dto"

type Tags struct {
	*dto.Page
}

type Tag struct {
	*dto.Tag
	Posts []*Post `json:"posts"`
}
