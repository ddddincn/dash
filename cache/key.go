package cache

import (
	"dash/consts"
	"strconv"
)

func BuildTokenAccessKey(accessToken string) string {
	return consts.TokenAccessCachePrefix + accessToken
}

func BuildTokenRefreshKey(refreshToken string) string {
	return consts.TokenRefreshCachePrefix + refreshToken
}

func BuildAccessTokenKey(userID int32) string {
	return consts.TokenAccessCachePrefix + strconv.Itoa(int(userID))
}

func BuildRefreshTokenKey(userID int32) string {
	return consts.TokenRefreshCachePrefix + strconv.Itoa(int(userID))
}

func BuildTokenBlacklistKey(tokenStr string) string {
	return consts.TokenBlacklistCachePrefix + tokenStr
}
