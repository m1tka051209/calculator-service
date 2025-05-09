package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

func StartHTTPGateway() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/register", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Login    string `json:"login"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
	})

	mux.HandleFunc("POST /api/v1/login", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Login    string `json:"login"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}
		
		token, _ := GenerateJWT("user123") // Замените на реальный ID пользователя
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	})

	mux.HandleFunc("POST /api/v1/calculate", func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if _, err := ValidateJWT(token); err != nil {
			respondError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		var req struct {
			Expression string `json:"expression"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"expression_id": "generated-id-123",
			"status":        "pending",
		})
	})

	mux.HandleFunc("GET /api/v1/expressions", func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if _, err := ValidateJWT(token); err != nil {
			respondError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{
			map[string]interface{}{
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

func respondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}