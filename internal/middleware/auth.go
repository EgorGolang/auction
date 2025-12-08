package middleware

import (
	"auction/internal/models"
	"auction/internal/pkg"
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"log"
	"net/http"
	"os"
	"strings"
)

const UserKey = "user"

func AuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Println("no authorization header")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			log.Println("authorization token is empty")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims := &pkg.CustomClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil {
			log.Printf("ERROR parsing token: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			log.Println("Token invalid token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		log.Printf("user authenticated - ID: %d, Username: %s, Role: %s",
			claims.UserID, claims.Username, claims.Role)

		user := &models.User{
			ID:       claims.UserID,
			Username: claims.Username,
			Email:    claims.Email,
			Role:     claims.Role,
		}

		ctx := context.WithValue(r.Context(), UserKey, user)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(UserKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}
