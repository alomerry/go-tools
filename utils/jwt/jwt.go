package jwt

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(claim CustomClaims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	tokenString, err := token.SignedString([]byte(secret))

	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func VerifyToken(tokenString string, secret string) (*CustomClaims, error) {
	var (
		claim = CustomClaims{}
	)
	token, err := jwt.ParseWithClaims(tokenString, &claim, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if ok && !token.Valid {
		return nil, errors.New("token is invalid")
	}

	if ok {
		return claims, nil
	}

	return nil, errors.New("unknown claims type, cannot verify token")
}
