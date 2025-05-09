package worker

import (
	"context"
	"testing"
	"time"

	"github.com/m1tka051209/calculator-service/server"
	// "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/m1tka051209/calculator-service/models"
)

type MockTaskManager struct {
	mock.Mock
}

func (m *MockTaskManager) GetNextTask() (*models.Task, error) {
	args := m.Called()
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskManager) UpdateTaskStatus(ctx context.Context, taskID, status string) error {
	args := m.Called(ctx, taskID, status)
	return args.Error(0)
}

func (m *MockTaskManager) UpdateTaskResult(ctx context.Context, taskID string, result float64) error {
	args := m.Called(ctx, taskID, result)
	return args.Error(0)
}

type MockCalculatorClient struct {
	mock.Mock
}

func (m *MockCalculatorClient) Calculate(ctx context.Context, req *server.CalculationRequest) (*server.CalculationResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*server.CalculationResponse), args.Error(1)
}

func TestProcessTasks(t *testing.T) {
	mockTM := new(MockTaskManager)
	mockClient := new(MockCalculatorClient)

	task := &models.Task{
		ID:        "task1",
		Arg1:      2,
		Arg2:      3,
		Operation: "+",
	}

	mockTM.On("GetNextTask").Return(task, nil)
	mockTM.On("UpdateTaskResult", mock.Anything, "task1", 5.0).Return(nil)
	mockClient.On("Calculate", mock.Anything, &server.CalculationRequest{
		Arg1:      2,
		Arg2:      3,
		Operation: "+",
	}).Return(&server.CalculationResponse{Result: 5}, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	processTasks(ctx, mockTM, mockClient, 1)

	mockTM.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}