package model

import "github.com/golang-jwt/jwt/v5"

type JwtCustomClaims struct {
	ID   int
	Name string
	jwt.RegisteredClaims
}
