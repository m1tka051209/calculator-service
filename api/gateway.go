package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

func StartHTTPGateway() http.Handler {
	mux := http.NewServeMux()

	// Регистрация
	mux.HandleFunc("POST /api/v1/register", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Login    string `json:"login"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}

		// Здесь должна быть логика регистрации через ваш репозиторий
		respondJSON(w, http.StatusOK, map[string]string{"status": "OK"})
	})

	// Авторизация
	mux.HandleFunc("POST /api/v1/login", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Login    string `json:"login"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}

		// Здесь должна быть проверка логина/пароля
		token, _ := GenerateJWT("user123") // Замените на реальный ID
		respondJSON(w, http.StatusOK, map[string]string{"token": token})
	})

	// Вычисление выражения
	mux.HandleFunc("POST /api/v1/calculate", func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if _, err := ValidateJWT(token); err != nil {
			respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}

		var req struct {
			Expression string `json:"expression"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}

		// Здесь вызов вашего gRPC сервиса
		respondJSON(w, http.StatusAccepted, map[string]string{
			"expression_id": "generated-id-123",
			"status":        "pending",
		})
	})

	// Получение выражений
	mux.HandleFunc("GET /api/v1/expressions", func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if _, err := ValidateJWT(token); err != nil {
			respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}

		// Здесь получение выражений из репозитория
		respondJSON(w, http.StatusOK, []map[string]interface{}{
			{
				"id":          "generated-id-123",
				"expression":  "2+2*2",
				"status":      "completed",
				"result":      6,
				"created_at":  "2025-05-09T16:20:00Z",
				"completed_at": "2025-05-09T16:20:05Z",
			},
		})
	})

	return mux
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}