package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	Category string `json:"category"`
	Id       string `json:"id"`
	Issuer   string `json:"issuer"`
	jwt.RegisteredClaims
}

func NewCustomClaims(category, id, issuer, validDuration string) CustomClaims {
	duration, err := time.ParseDuration(validDuration)
	if err != nil {
		duration = time.Hour * 24
	}

	now := time.Now()
	return CustomClaims{
		Category: category,
		Id:       id,
		Issuer:   issuer,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
}
