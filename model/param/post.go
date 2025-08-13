package param

import (
	"dash/consts"
)

type Post struct {
	Title           string             `json:"title" form:"title" binding:"gte=1,lte=100"`
	Status          consts.PostStatus  `json:"status" form:"status" binding:"gte=0"`
	Slug            string             `json:"slug" form:"slug" binding:"lte=255"`
	EditorType      *consts.EditorType `json:"editor_type" form:"editor_type"`
	Content         string             `json:"content" form:"content"`
	OriginalContent string             `json:"original_content" form:"original_content"`
	Summary         string             `json:"summary" form:"summary"`
	Thumbnail       string             `json:"thumbnail" form:"thumbnail"`
	TopPriority     int32              `json:"top_priority" form:"top_priority" binding:"gte=0"`
	TagIDs          []int32            `json:"tag_ids" form:"tag_ids"`
	CategoryIDs     []int32            `json:"category_ids" form:"category_ids"`
}

type PostContent struct {
	Content         string `json:"content" form:"content"`
	OriginalContent string `json:"original_content" form:"original_content"`
}

type PostQuery struct {
	Page
	*Sort
	Keyword    *string              `json:"keyword" form:"keyword"`
	Statuses   []*consts.PostStatus `json:"statuses" form:"statuses"`
	CategoryID *int32               `json:"category_id" form:"category_id"`
	Detail     *bool                `json:"detail" form:"detail"`
	TagID      *int32               `json:"tag_id" form:"tag_id"`
	// WithPassword *bool                `json:"-" form:"-"`
}
