package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"dash/config"
	"dash/utils/xerr"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisCache 单机Redis缓存
type RedisCache struct {
	db         *redis.Client
	defaultTTL time.Duration
}

var Cache *RedisCache
var Redis *redis.Client

// NewRedisCache 创建一个新的 Redis 缓存实例
func NewRedisCache(conf *config.Config, logger *zap.Logger) *RedisCache {
	client, err := initRedis(conf, logger)
	if err != nil {
		logger.Fatal("connect to Redis error", zap.Error(err))
		return nil
	}
	if err := client.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("connect to Redis error", zap.Error(err))
		return nil
	}
	Redis = client
	Cache = &RedisCache{
		db:         client,
		defaultTTL: conf.Cache.DefaultTTL,
	}
	return Cache
}

func initRedis(conf *config.Config, logger *zap.Logger) (*redis.Client, error) {
	redisConfig := conf.Cache.Redis
	if redisConfig == nil {
		return nil, xerr.WithMsg(nil, "nil Redis config")
	}

	logger.Info("try connect to Redis ", zap.Any("addr", redisConfig))

	client := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})

	return client, nil
}

func SetDefault(key string, value interface{}) error {
	return Set(key, value, Cache.defaultTTL)
}

func Set(key string, value interface{}, ttl time.Duration) error {
	// 将值序列化为 JSON
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return Cache.db.Set(context.Background(), key, bytes, ttl).Err()
}

func Get(key string) (interface{}, bool, error) {
	var result interface{}

	val, err := Cache.db.Get(context.Background(), key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// 缓存未命中，这不是一个真正的错误
			return result, false, nil
		}
		return result, false, err
	}

	// 反序列化 JSON 到泛型类型
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return result, false, err
	}

	return result, true, nil
}

func Delete(key string) error {
	return Cache.db.Del(context.Background(), key).Err()
}

func BatchDelete(keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	return Cache.db.Del(context.Background(), keys...).Err()
}
