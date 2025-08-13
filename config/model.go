package config

import "time"

type Config struct {
	Server     Server      `mapstructure:"server" json:"server"`
	Log        Log         `mapstructure:"logging" json:"logging"`
	PostgreSQL *PostgreSQL `mapstructure:"postgre" json:"postgre"`
	MySQL      *MySQL      `mapstructure:"mysql" json:"mysql"`
	Cache      *Cache      `mapstructure:"cache" json:"cache"`
	SQLite3    *SQLite3    `mapstructure:"sqlite3" json:"sqlite3"`
	Dash       Dash        `mapstructure:"dash" json:"dash"`
}

type PostgreSQL struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     string `mapstructure:"port" json:"port"`
	DB       string `mapstructure:"db" json:"db"`
	Username string `mapstructure:"username" json:"username"`
	Password string `mapstructure:"password" json:"password"`
}

type MySQL struct {
	Dsn string `mapstructure:"dsn" json:"dsn"`
}

type Cache struct {
	Redis      *Redis        `mapstructure:"redis" json:"redis"`
	DefaultTTL time.Duration `mapstructure:"default_ttl" json:"default_ttl"`
}

type Redis struct {
	Addr     string `mapstructure:"addr" json:"addr"`
	Password string `mapstructure:"password" json:"password"`
	DB       int    `mapstructure:"db" json:"db"`
}

type SQLite3 struct {
	Enable bool `mapstructure:"enable" json:"enable"`
	File   string
}

type Server struct {
	Host string `mapstructure:"host" json:"host"`
	Port string `mapstructure:"port" json:"port"`
}

type Log struct {
	FileName string `mapstructure:"filename" json:"filename"`
	Levels   Levels `mapstructure:"level" json:"level"`
	MaxSize  int    `mapstructure:"max_size" json:"max_size"`
	MaxAge   int    `mapstructure:"max_age" json:"max_age"`
	Compress bool   `mapstructure:"compress" json:"compress"`
}

type Levels struct {
	App  string `mapstructure:"app" json:"app"`
	Gorm string `mapstructure:"gorm" json:"gorm"`
}

type LogMode string

const (
	Console LogMode = "console"
	File    LogMode = "file"
)

type Dash struct {
	Mode              string  `mapstructure:"mode"`
	LogMode           LogMode `mapstructure:"log_mode"`
	WorkDir           string  `mapstructure:"work_dir"`
	UploadDir         string
	LogDir            string `mapstructure:"log_dir"`
	TemplateDir       string `mapstructure:"template_dir"`
	ThemeDir          string
	AdminResourcesDir string
	AdminURLPath      string `mapstructure:"admin_url_path"`
}
