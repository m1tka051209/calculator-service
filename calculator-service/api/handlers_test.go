package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/m1tka051209/calculator-service/api"
	"github.com/m1tka051209/calculator-service/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) AssertExpectations(t *testing.T) {
	panic("unimplemented")
}

func (m *MockRepository) CreateUser(ctx context.Context, login, passwordHash string) error {
	args := m.Called(ctx, login, passwordHash)
	return args.Error(0)
}

func TestRegisterHandler_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockRepo.On("CreateUser", mock.Anything, "test", mock.Anything).Return(nil)

	h := api.NewHandlers(mockRepo, "secret", 24*time.Hour)

	req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"login":"test","password":"123"}`))
	w := httptest.NewRecorder()

	h.RegisterHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}
