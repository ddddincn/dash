package vo

import (
	"dash/model/dto"
)

type Post struct {
	dto.Post
	Tags       []*dto.Tag      `json:"tags"`
	Categories []*dto.Category `json:"categories"`
}

type PostDetail struct {
	dto.PostDetail
	Tags       []*dto.Tag       `json:"tags"`
	Categories []*dto.Category  `json:"categories"`
	PrePost    *dto.PostOutline `json:"pre_post"`
	NextPost   *dto.PostOutline `json:"next_post"`
}
