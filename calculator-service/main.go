package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/m1tka051209/calculator-service/api"
	"github.com/m1tka051209/calculator-service/calculator"
	"github.com/m1tka051209/calculator-service/config"
	"github.com/m1tka051209/calculator-service/db"
	"github.com/m1tka051209/calculator-service/task_manager"
)

func main() {
	cfg := config.Load()

	repo, err := db.NewSQLiteRepository(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запуск менеджера задач
	tm := task_manager.NewTaskManager(repo)
	go taskWorker(ctx, tm, cfg.WorkerPoolSize)

	// Настройка HTTP сервера
	handlers := api.NewHandlers(repo, cfg.JWTSecret, cfg.TokenExpiration)
	router := http.NewServeMux()
	router.HandleFunc("/api/v1/register", handlers.RegisterHandler)
	router.HandleFunc("/api/v1/login", handlers.LoginHandler)
	router.Handle("/api/v1/calculate", api.AuthMiddleware(handlers, http.HandlerFunc(handlers.CalculateHandler)))

	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
		cancel() // Остановка воркеров
	}()

	log.Printf("Server started on :%s", cfg.HTTPPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

func taskWorker(ctx context.Context, tm *task_manager.TaskManager, poolSize int) {
	for i := 0; i < poolSize; i++ {
		go func(workerID int) {
			for {
				select {
				case <-ctx.Done():
					log.Printf("Worker %d shutting down", workerID)
					return
				default:
					task, err := tm.GetNextTask()
					if err != nil {
						log.Printf("Worker %d error getting task: %v", workerID, err)
						time.Sleep(2 * time.Second)
						continue
					}

					if task == nil {
						time.Sleep(1 * time.Second)
						continue
					}

					result := calculator.Calculate(task)
					if err := tm.SaveTaskResult(task.ID, result); err != nil {
						log.Printf("Worker %d error saving result: %v", workerID, err)
					}
				}
			}
		}(i)
	}
}
