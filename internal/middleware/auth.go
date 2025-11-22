package middleware

import (
	"auction/internal/models"
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
			log.Println("No Authorization header")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			log.Println("Authorization header valid")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil {
			log.Printf("ERROR parsing token", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Printf("ERROR: Expecred MapClaims")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userUDFloat, ok := claims["user_id"].(float64)
		if !ok {
			log.Printf("ERROR: user_id not found or wrong type")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		username, _ := claims["username"].(string)
		email, _ := claims["email"].(string)
		role, _ := claims["role"].(string)

		user := &models.User{
			ID:       int(userUDFloat),
			Username: username,
			Email:    email,
			Role:     role,
		}

		ctx := context.WithValue(r.Context(), UserKey, user)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(UserKey)
		if user == nil {
			log.Println("RequireAuth user not found")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserKey, user)))
	})
}

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(UserKey).(*models.User)
			if !ok || user.Role != role {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func GetUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(UserKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}
