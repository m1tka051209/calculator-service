package api

import (
    "context"
    "github.com/m1tka051209/calculator-service/models"
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

func (m *MockRepository) GetExpressionsByUser(ctx context.Context, userID string) ([]models.Expression, error) {
    args := m.Called(ctx, userID)
    return args.Get(0).([]models.Expression), args.Error(1)
}

func (m *MockRepository) GetPendingTasks(ctx context.Context, limit int) ([]models.Task, error) {
    args := m.Called(ctx, limit)
    return args.Get(0).([]models.Task), args.Error(1)
}

func (m *MockRepository) UpdateTaskResult(ctx context.Context, taskID string, result float64) error {
    args := m.Called(ctx, taskID, result)
    return args.Error(0)
}

func (m *MockRepository) UpdateTaskStatus(ctx context.Context, taskID, status string) error {
    args := m.Called(ctx, taskID, status)
    return args.Error(0)
}

func (m *MockRepository) Close() error {
    args := m.Called()
    return args.Error(0)
}