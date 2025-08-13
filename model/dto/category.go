package dto

type Category struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
	CreateTime  int64  `json:"create_time"`
	FullPath    string `json:"full_path"`
	Priority    int32  `json:"priority"`
}

type CategoryWithPostCount struct {
	*Category
	PostCount int64 `json:"post_count"`
}
