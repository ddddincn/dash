package vo

type User struct {
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	Description string `json:"description"`
}

type Settings struct {
	Icon         string `json:"icon"`
	AvatarCircle bool   `json:"avatar_circle"`
	SidebarWidth string `json:"sidebar_width"`
	RSS          string `json:"rss"`
	Twitter      string `json:"twitter"`
	Facebook     string `json:"facebook"`
	Instagram    string `json:"instagram"`
	Weibo        string `json:"weibo"`
	QQ           string `json:"qq"`
	Telegram     string `json:"telegram"`
	Email        string `json:"email"`
	Github       string `json:"github"`
}

type SidebarInfo struct {
	BlogURL   string    `json:"blog_url"`
	BlogTitle string    `json:"blog_title"`
	User      *User     `json:"user"`
	Settings  *Settings `json:"settings"`
	// Menus         []*dto.Menu             `json:"menus"`
	// Tags          []*dto.TagWithPostCount `json:"tags"`
}
