package param

type Install struct {
	User
	Database Database `json:"database"`
	Redis    Redis    `json:"redis"`
	Title    string   `json:"title" binding:"required"`
	URL      string   `json:"url"`
}

type Database struct {
	Type     string  `json:"type"  binding:"required"`
	Host     *string `json:"host"`
	Port     *int    `json:"port"`
	Database *string `json:"database"`
	Username *string `json:"username"`
	Password *string `json:"password"`
}

type Redis struct {
	Host     string  `json:"host"`
	Port     int     `json:"port"`
	Database int     `json:"database"`
	Username *string `json:"username"`
	Password *string `json:"password"`
}
