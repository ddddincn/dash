package dto

type Tag struct {
	ID         int32  `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	Thumbnail  string `json:"thumbnail"`
	CreateTime int64  `json:"create_time"`
	FullPath   string `json:"full_path"`
	Color      string `json:"color"`
}

type TagWithPostCount struct {
	*Tag
	PostCount int64 `json:"post_count"`
}
