package calculator

import (
	"testing"
	"github.com/m1tka051209/calculator-service/models"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		name     string
		task     models.Task
		expected float64
	}{
		{"addition", models.Task{Arg1: 2, Arg2: 3, Operation: "+"}, 5},
		{"division by zero", models.Task{Arg1: 5, Arg2: 0, Operation: "/"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Calculate(&tt.task); got != tt.expected {
				t.Errorf("Calculate() = %v, want %v", got, tt.expected)
			}
		})
	}
}