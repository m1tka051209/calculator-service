package main

import (
	// "context"
	"log"
	"time"
	"github.com/m1tka051209/calculator-service/config"
	"github.com/m1tka051209/calculator-service/db"
	"github.com/m1tka051209/calculator-service/models"
	"github.com/m1tka051209/calculator-service/task_manager"
)

func main() {
	cfg := config.Load()
	
	repo, err := db.NewSQLiteRepository(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	tm := task_manager.NewTaskManager(repo)

	for {
		task, err := tm.GetNextTask()
		if err != nil {
			log.Printf("Error getting task: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if task == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		result := calculate(task)
		if err := tm.SaveTaskResult(task.ID, result); err != nil {
			log.Printf("Error saving result: %v", err)
		}
	}
}

func calculate(task *models.Task) float64 {
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