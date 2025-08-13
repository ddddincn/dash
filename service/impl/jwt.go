package impl

import (
	"context"
	"dash/cache"
	"dash/consts"
	"dash/model"
	"dash/model/entity"
	"dash/model/property"
	"dash/service"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type jwtServiceImpl struct {
	OptionService service.OptionService
	accessSecret  []byte
	refreshSecret []byte
}

func NewJWTService(optionService service.OptionService) service.JWTService {
	ctx := context.Background()
	access := optionService.GetOrByDefault(ctx, property.JWTAccessSecret)
	refresh := optionService.GetOrByDefault(ctx, property.JWTRefreshSecret)

	return &jwtServiceImpl{
		OptionService: optionService,
		accessSecret:  []byte(access.(string)),
		refreshSecret: []byte(refresh.(string)),
	}
}

func (j *jwtServiceImpl) GenerateTokens(user *entity.User) (string, string, error) {
	err := j.cleanOldTokens(user.ID)
	if err != nil {
		return "", "", err
	}
	iJwtCustomClaims := &model.JwtCustomClaims{
		ID:   int(user.ID),
		Name: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ID: strconv.Itoa(int(user.ID)),
			// 设置过期时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(consts.AccessTokenExpiredSeconds * time.Second)),
			// 颁发时间
			IssuedAt: jwt.NewNumericDate(time.Now()),
			// 发布者
			Issuer: "Dash",
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, iJwtCustomClaims)
	accessTokenStr, err := accessToken.SignedString(j.accessSecret)
	if err != nil {
		return "", "", err
	}

	iJwtCustomClaims = &model.JwtCustomClaims{
		ID:   int(user.ID),
		Name: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ID: strconv.Itoa(int(user.ID)),
			// 设置过期时间 在当前基础上 添加一个小时后 过期
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(consts.RefreshTokenExpiredDays * time.Hour * 24)),
			// 颁发时间 也就是生成时间
			IssuedAt: jwt.NewNumericDate(time.Now()),
			//主题
			Issuer: "Dash",
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, iJwtCustomClaims)
	refreshTokenStr, err := refreshToken.SignedString(j.refreshSecret)
	if err != nil {
		return "", "", err
	}

	ctx := context.Background()
	pipe := cache.Redis.TxPipeline()

	pipe.Set(ctx, cache.BuildTokenAccessKey(accessTokenStr), user.ID, time.Second*consts.AccessTokenExpiredSeconds)
	pipe.Set(ctx, cache.BuildTokenRefreshKey(refreshTokenStr), user.ID, 24*time.Hour*consts.RefreshTokenExpiredDays)

	pipe.Set(ctx, cache.BuildAccessTokenKey(user.ID), accessTokenStr, time.Second*consts.AccessTokenExpiredSeconds)
	pipe.Set(ctx, cache.BuildRefreshTokenKey(user.ID), refreshTokenStr, 24*time.Hour*consts.RefreshTokenExpiredDays)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return "", "", err
	}
	return accessTokenStr, refreshTokenStr, nil
}

func (j *jwtServiceImpl) ParseAccessToken(tokenStr string) (*model.JwtCustomClaims, error) {
	claims, err := j.parseToken(tokenStr, &model.JwtCustomClaims{}, j.accessSecret) // 解析 Token
	if err != nil {
		return nil, err
	}
	if customClaims, ok := claims.(*model.JwtCustomClaims); ok { // 确保解析出的 Claims 类型正确
		return customClaims, nil
	}
	return nil, errors.New("invalid token")
}

func (j *jwtServiceImpl) ParseRefreshToken(tokenStr string) (*model.JwtCustomClaims, error) {
	claims, err := j.parseToken(tokenStr, &model.JwtCustomClaims{}, j.refreshSecret) // 解析 Token
	if err != nil {
		return nil, err
	}
	if customClaims, ok := claims.(*model.JwtCustomClaims); ok { // 确保解析出的 Claims 类型正确
		return customClaims, nil
	}
	return nil, errors.New("invalid token")
}

func (j *jwtServiceImpl) parseToken(tokenStr string, claims jwt.Claims, secretKey interface{}) (interface{}, error) {
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return token.Claims, nil
}

func (j *jwtServiceImpl) cleanOldTokens(userID int32) error {
	ctx := context.Background()
	accessTokenKeyByID := cache.BuildAccessTokenKey(userID)
	refreshTokenKeyByID := cache.BuildRefreshTokenKey(userID)
	accessToken, err := cache.Redis.Get(ctx, accessTokenKeyByID).Result()
	ok := true
	if err != nil {
		if errors.Is(err, redis.Nil) {
			ok = false
		} else {
			return err
		}
	}
	if ok {
		err = cache.Redis.Del(ctx, accessTokenKeyByID, cache.BuildTokenAccessKey(accessToken)).Err()
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				return err
			}
		}
	}
	refreshToken, err := cache.Redis.Get(ctx, refreshTokenKeyByID).Result()
	ok = true
	if err != nil {
		if errors.Is(err, redis.Nil) {
			ok = false
		} else {
			return err
		}
	}
	if ok {
		err = cache.Redis.Del(ctx, refreshTokenKeyByID, cache.BuildTokenRefreshKey(refreshToken)).Err()
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				return err
			}
		}
	}
	return nil
}

func (j *jwtServiceImpl) RefreshToken(refreshToken string) (string, error) {

	// 解析refresh token获取用户信息
	claims, err := j.ParseRefreshToken(refreshToken)
	if err != nil {
		return "", errors.New("invalid refresh token: " + err.Error())
	}

	// 检查refresh token是否已过期
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return "", errors.New("refresh token has expired")
	}

	// 创建新的access token claims
	newClaims := &model.JwtCustomClaims{
		ID:   claims.ID,   // 使用原有的用户ID
		Name: claims.Name, // 使用原有的用户名
		RegisteredClaims: jwt.RegisteredClaims{
			ID: strconv.Itoa(int(claims.ID)),
			// 设置过期时间 在当前基础上 添加一个小时后 过期
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(consts.AccessTokenExpiredSeconds * time.Second)),
			// 颁发时间 也就是生成时间
			IssuedAt: jwt.NewNumericDate(time.Now()),
			//主题
			Issuer: "Dash",
		},
	}

	// 生成新的access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	accessTokenStr, err := accessToken.SignedString(j.accessSecret)
	if err != nil {
		return "", errors.New("failed to generate new access token: " + err.Error())
	}
	ctx := context.Background()
	pipe := cache.Redis.TxPipeline()
	oldToken, err := pipe.Get(ctx, cache.BuildAccessTokenKey(int32(claims.ID))).Result()
	if !errors.Is(err, redis.Nil) {
		pipe.Del(ctx, cache.BuildTokenAccessKey(oldToken))
	}
	pipe.Set(ctx, cache.BuildTokenAccessKey(accessTokenStr), claims.ID, time.Second*consts.AccessTokenExpiredSeconds)
	pipe.Set(ctx, cache.BuildAccessTokenKey(int32(claims.ID)), accessTokenStr, time.Second*consts.AccessTokenExpiredSeconds)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return "", errors.New("failed to generate new access token: " + err.Error())
	}
	return accessTokenStr, nil
}

func (j *jwtServiceImpl) IsInCache(token string) bool {
	_, err := cache.Redis.Get(context.Background(), cache.BuildTokenAccessKey(token)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			_, err = cache.Redis.Get(context.Background(), cache.BuildTokenRefreshKey(token)).Result()
			if err != nil {
				if errors.Is(err, redis.Nil) {
					return false
				}

				return false
			}
		} else {
			return false
		}
	}

	return true
}
