package task_manager

import (
    "context"
    "github.com/m1tka051209/calculator-service/db"
    "github.com/m1tka051209/calculator-service/models"
)

type TaskManager struct {
    repo *db.Repository
}

func NewTaskManager(repo *db.Repository) *TaskManager {
    return &TaskManager{repo: repo}
}

func (tm *TaskManager) GetNextTask() (*models.Task, bool) {
    tasks, err := tm.repo.GetPendingTasks(context.Background(), 1)
    if err != nil || len(tasks) == 0 {
        return nil, false
    }
    return &tasks[0], true
}

func (tm *TaskManager) SaveTaskResult(taskID string, result float64) error {
    return tm.repo.UpdateTaskResult(context.Background(), taskID, result)
}