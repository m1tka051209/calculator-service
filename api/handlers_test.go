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

func TestCalculateHandler_Valid(t *testing.T) {
    mockRepo := &MockRepository{}
    mockRepo.On("CreateExpression", mock.Anything, "user123", "2+2").Return("expr123", nil)

    h := &Handlers{
        repo:          mockRepo,
        jwtSecret:     "secret",
        jwtExpiration: 24 * time.Hour,
    }

    req := httptest.NewRequest("POST", "/calculate", strings.NewReader(`{"expression":"2+2"}`))
    req.Header.Set("Content-Type", "application/json")
    
    ctx := context.WithValue(req.Context(), UserIDKey, "user123")
    req = req.WithContext(ctx)
    
    w := httptest.NewRecorder()
    h.CalculateHandler(w, req)

    assert.Equal(t, http.StatusAccepted, w.Code)
    assert.Contains(t, w.Body.String(), "expr123")
    mockRepo.AssertExpectations(t)
}

func (m *MockRepository) GetPendingTasks(ctx context.Context, limit int) ([]models.Task, error) {
    args := m.Called(ctx, limit)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]models.Task), args.Error(1)
}

func (m *MockRepository) UpdateTaskResult(ctx context.Context, taskID string, result float64) error {
    args := m.Called(ctx, taskID, result)
    return args.Error(0)
}

func TestRegisterHandler_Success(t *testing.T) {
    mockRepo := &MockRepository{}
    mockRepo.On("CreateUser", mock.Anything, "test", mock.Anything).Return(nil)

    h := &Handlers{
        repo:          mockRepo,
        jwtSecret:     "secret",
        jwtExpiration: 24 * time.Hour,
    }

    req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"login":"test","password":"123"}`))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    h.RegisterHandler(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    mockRepo.AssertExpectations(t)
}

func TestLoginHandler_Success(t *testing.T) {
    mockRepo := &MockRepository{}
    user := &models.User{
        ID:           "123",
        Login:        "test",
        PasswordHash: "$2a$10$fakehash",
    }
    mockRepo.On("GetUserByLogin", mock.Anything, "test").Return(user, nil)

    h := &Handlers{
        repo:          mockRepo,
        jwtSecret:     "secret",
        jwtExpiration: 24 * time.Hour,
    }

    req := httptest.NewRequest("POST", "/login", strings.NewReader(`{"login":"test","password":"123"}`))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    h.LoginHandler(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), "token")
    mockRepo.AssertExpectations(t)
}