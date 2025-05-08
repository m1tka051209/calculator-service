package api

import (
	"time"
	"github.com/golang-jwt/jwt/v5"
)

func (h *Handlers) GenerateToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(h.jwtExpiration).Unix(),
	})
	return token.SignedString([]byte(h.jwtSecret))
}