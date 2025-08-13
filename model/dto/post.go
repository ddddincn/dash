package dto

import "dash/consts"

type PostOutline struct {
	ID         int32             `json:"id"`
	Title      string            `json:"title"`
	Status     consts.PostStatus `json:"status"`
	Slug       string            `json:"slug"`
	EditorType consts.EditorType `json:"editor_type"`
	CreateTime int64             `json:"create_time"`
	UpdateTime int64             `json:"update_time"`
	FullPath   string            `json:"full_path"`
}
type Post struct {
	PostOutline
	Summary     string `json:"summary"`
	Thumbnail   string `json:"thumbnail"`
	Visits      int64  `json:"visits"`
	TopPriority int32  `json:"top_priority"`
	Likes       int64  `json:"likes"`
	WordCount   int64  `json:"word_count"`
	Topped      bool   `json:"topped"`
}

type PostDetail struct {
	Post
	OriginalContent string `json:"original_content"`
	Content         string `json:"content"`
}
