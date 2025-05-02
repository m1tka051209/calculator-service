package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/m1tka051209/calculator-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type Repository interface {
	CreateUser(ctx context.Context, login, passwordHash string) error
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	CreateExpression(ctx context.Context, userID, expr string) (string, error)
	GetPendingTasks(ctx context.Context, limit int) ([]models.Task, error)
	UpdateTaskResult(ctx context.Context, taskID string, result float64) error
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateUser(ctx context.Context, login, passwordHash string) error {
	args := m.Called(ctx, login, passwordHash)
	return args.Error(0)
}

func (m *MockRepository) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	args := m.Called(ctx, login)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockRepository) CreateExpression(ctx context.Context, userID, expr string) (string, error) {
	args := m.Called(ctx, userID, expr)
	return args.String(0), args.Error(1)
}

func (m *MockRepository) GetPendingTasks(ctx context.Context, limit int) ([]models.Task, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]models.Task), args.Error(1)
}

func (m *MockRepository) UpdateTaskResult(ctx context.Context, taskID string, result float64) error {
	args := m.Called(ctx, taskID, result)
	return args.Error(0)
}

func TestRegisterHandler_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockRepo.On("CreateUser", mock.Anything, "test", mock.Anything).Return(nil)

	h := NewHandlers(mockRepo, "secret", 24*time.Hour)

	req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"login":"test","password":"123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.RegisterHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}