package pkg

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"os"
	"time"
)

type CustomClaims struct {
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func GenerateToken(userID int, username, email, role, tokenType string) (string, error) {
	var expirationTime time.Time

	if tokenType == "access" {
		expirationTime = time.Now().Add(1 * time.Hour)
	} else if tokenType == "refresh" {
		expirationTime = time.Now().Add(7 * 24 * time.Hour)
	}

	claims := &CustomClaims{
		UserID:    userID,
		Username:  username,
		Email:     email,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func ValidateToken(tokenStr string) (*CustomClaims, error) {
	claims := &CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
