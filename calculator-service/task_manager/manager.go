package task_manager

import (
	"context"
	"log"
	"time"

	"github.com/m1tka051209/calculator-service/db"
	"github.com/m1tka051209/calculator-service/models"
)

type TaskManager struct {
	repo db.Repository
}

func NewTaskManager(repo db.Repository) *TaskManager {
	return &TaskManager{repo: repo}
}

func (tm *TaskManager) GetNextTask() (*models.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tasks, err := tm.repo.GetPendingTasks(ctx, 1)
	if err != nil {
		log.Printf("Error getting tasks: %v", err)
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, nil
	}

	return &tasks[0], nil
}

func (tm *TaskManager) SaveTaskResult(taskID string, result float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return tm.repo.UpdateTaskResult(ctx, taskID, result)
}