package dto

type Statistic struct {
	PostCount     int64 `json:"post_count"`
	CategoryCount int64 `json:"category_count"`
	TagCount      int64 `json:"tag_count"`
	Birthday      int64 `json:"birthday"`
	EstablishDays int64 `json:"establish_days"`
	VisitCount    int64 `json:"visit_count"`
	LikeCount     int64 `json:"like_count"`
}
