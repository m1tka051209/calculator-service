package calculator

import (
	"github.com/m1tka051209/calculator-service/models"
	"time"
)

func Calculate(task *models.Task) float64 {
	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)
	
	switch task.Operation {
	case "+":
		return task.Arg1 + task.Arg2
	case "-":
		return task.Arg1 - task.Arg2
	case "*":
		return task.Arg1 * task.Arg2
	case "/":
		if task.Arg2 == 0 {
			return 0
		}
		return task.Arg1 / task.Arg2
	default:
		return 0
	}
}

func ValidateOperation(op string) bool {
	switch op {
	case "+", "-", "*", "/":
		return true
	default:
		return false
	}
}