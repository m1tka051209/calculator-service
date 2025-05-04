package main

import (
	"log"
	"os"
	"time"

	"github.com/m1tka051209/calculator-service/calculator"
	"github.com/m1tka051209/calculator-service/config"
	"github.com/m1tka051209/calculator-service/db"
	"github.com/m1tka051209/calculator-service/task_manager"
)

func init() {
	logFile, err := os.OpenFile("worker.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFile)
	}
}

func main() {
	log.Println("Starting worker...")
	cfg := config.Load()
	
	repo, err := db.NewSQLiteRepository(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	tm := task_manager.NewTaskManager(repo)
	log.Printf("Worker initialized with DB path: %s", cfg.DBPath)

	for {
		log.Println("Checking for new tasks...")
		task, err := tm.GetNextTask()
		if err != nil {
			log.Printf("Error getting task: %v (retrying in 2s)", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if task == nil {
			log.Println("No tasks available (retrying in 1s)")
			time.Sleep(1 * time.Second)
			continue
		}

		log.Printf("Processing task ID: %s, Operation: %f %s %f", 
			task.ID, task.Arg1, task.Operation, task.Arg2)
		
		result := calculator.Calculate(task)
		log.Printf("Task result: %f", result)

		if err := tm.SaveTaskResult(task.ID, result); err != nil {
			log.Printf("Error saving result: %v", err)
		}
	}
}