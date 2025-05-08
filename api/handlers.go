package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/m1tka051209/calculator-service/calculator"
	"github.com/m1tka051209/calculator-service/db"
	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
)

type Handlers struct {
	repo          db.Repository
	jwtSecret     string
	jwtExpiration time.Duration
}

func NewHandlers(repo db.Repository, jwtSecret string, jwtExpiration time.Duration) *Handlers {
	return &Handlers{
		repo:          repo,
		jwtSecret:     jwtSecret,
		jwtExpiration: jwtExpiration,
	}
}

func (h *Handlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	if err := h.repo.CreateUser(r.Context(), req.Login, string(hashedPassword)); err != nil {
		respondError(w, http.StatusConflict, "user already exists")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
}

func (h *Handlers) GenerateToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(h.jwtExpiration).Unix(),
	})
	return token.SignedString([]byte(h.jwtSecret))
}

func (h *Handlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request")
		return
	}

	user, err := h.repo.GetUserByLogin(r.Context(), req.Login)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	tokenString, err := h.GenerateToken(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func (h *Handlers) CalculateHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		respondError(w, http.StatusUnauthorized, "user ID missing")
		return
	}

	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if !calculator.ValidateExpression(req.Expression) {
		respondError(w, http.StatusBadRequest, "invalid expression format")
		return
	}

	exprID, err := h.repo.CreateExpression(r.Context(), userID, req.Expression)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create expression")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"expression_id": exprID,
		"status":        "pending",
	})
}

func (h *Handlers) GetExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		respondError(w, http.StatusUnauthorized, "user ID missing")
		return
	}

	exprs, err := h.repo.GetExpressionsByUser(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get expressions")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exprs)
}

func respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
