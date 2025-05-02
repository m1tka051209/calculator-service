package main

import (
	"log"
	"net/http"
	"time"

	"github.com/m1tka051209/calculator-service/api"
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
	go func() {
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
	}()

	handlers := api.NewHandlers(repo, cfg.JWTSecret, cfg.TokenExpiration)

	http.HandleFunc("/api/v1/register", handlers.RegisterHandler)
	http.HandleFunc("/api/v1/login", handlers.LoginHandler)
	http.Handle("/api/v1/calculate", api.AuthMiddleware(handlers, http.HandlerFunc(handlers.CalculateHandler)))

	log.Printf("Server started on :%s", cfg.HTTPPort)
	log.Fatal(http.ListenAndServe(":"+cfg.HTTPPort, nil))
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