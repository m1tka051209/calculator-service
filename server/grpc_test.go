package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/mock"
)

// type mockRepository struct {
// 	mock.Mock
// }

// func (m *mockRepository) CreateExpression(ctx context.Context, userID, expr string) (string, error) {
// 	args := m.Called(ctx, userID, expr)
// 	return args.String(0), args.Error(1)
// }

func TestCalculate(t *testing.T) {
	server := &CalculatorServer{}

	tests := []struct {
		name     string
		request  *CalculationRequest
		expected float64
	}{
		{
			name: "simple addition",
			request: &CalculationRequest{
				Arg1:      2,
				Arg2:      3,
				Operation: "+",
			},
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.Calculate(context.Background(), tt.request)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, resp.Result)
		})
	}
}