package service

import (
	"dash/model"
	"dash/model/entity"
)

type JWTService interface {
	GenerateTokens(user *entity.User) (string, string, error)
	ParseAccessToken(tokenStr string) (*model.JwtCustomClaims, error)
	ParseRefreshToken(tokenStr string) (*model.JwtCustomClaims, error)
	// CleanOldTokens(userID int32) error
	// JoinBlackList(tokenStr string) error
	// IsInBlackList(tokenStr string) (bool, error)
	RefreshToken(refreshToken string) (string, error)
}
