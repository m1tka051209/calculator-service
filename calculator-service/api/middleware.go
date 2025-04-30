package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/m1tka051209/calculator-service/db"
)

// AuthMiddleware проверяет JWT токен и добавляет userID в контекст
func AuthMiddleware(repo *db.Repository, jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Извлекаем токен из заголовка
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondError(w, http.StatusUnauthorized, "Authorization header is required")
				return
			}

			// Проверяем формат "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondError(w, http.StatusUnauthorized, "Invalid authorization header format")
				return
			}

			tokenString := parts[1]

			// Парсим токен
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Проверяем алгоритм подписи
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil {
				respondError(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			// Проверяем claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok || !token.Valid {
				respondError(w, http.StatusUnauthorized, "Invalid token claims")
				return
			}

			// Извлекаем userID из токена
			userID, ok := claims["sub"].(string)
			if !ok {
				respondError(w, http.StatusUnauthorized, "Invalid user ID in token")
				return
			}

			// Проверяем существование пользователя
			_, err = repo.GetUserByID(r.Context(), userID)
			if err != nil {
				respondError(w, http.StatusUnauthorized, "User not found")
				return
			}

			// Добавляем userID в контекст запроса
			ctx := context.WithValue(r.Context(), "userID", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// respondError отправляет JSON с ошибкой
func respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}