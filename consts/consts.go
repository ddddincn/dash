package consts

const (
	DashUploadDir       = "upload"  //默认附件上传路径
	DashDefaultTagColor = "#cfd3d7" // 默认标签颜色
)

const (
	AccessTokenExpiredSeconds = 15 * 60
	RefreshTokenExpiredDays   = 1
	TokenAccessCachePrefix    = "admin_access_token_"
	TokenRefreshCachePrefix   = "admin_refresh_token_"
	TokenBlacklistCachePrefix = "token_blacklist_"
	OneTimeTokenQueryName     = "ott"

	AdminTokenHeaderName = "Authorization"
	AuthorizedUser       = "authorized_user"
)
