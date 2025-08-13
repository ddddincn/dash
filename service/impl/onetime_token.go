package impl

import (
	"context"
	"dash/cache"
	"dash/service"
	"dash/utils"
	"time"
)

const (
	oneTimeTokenPrefix = "OTT-"
	ottExpirationTime  = time.Minute * 5
)

type oneTimeTokenServiceImpl struct {
}

func NewOneTimeTokenService() service.OneTimeTokenService {
	return &oneTimeTokenServiceImpl{}
}

func (o *oneTimeTokenServiceImpl) Get(oneTimeToken string) (string, bool) {
	ctx := context.Background()
	v, err := cache.Redis.Get(ctx, oneTimeTokenPrefix+oneTimeToken).Result()
	if err != nil {
		return "", false
	}
	return v, true
}

func (o *oneTimeTokenServiceImpl) Create(value string) string {
	ctx := context.Background()
	uuid := utils.GenUUIDWithOutDash()
	cache.Redis.Set(ctx, oneTimeTokenPrefix+uuid, value, ottExpirationTime)
	return uuid
}
