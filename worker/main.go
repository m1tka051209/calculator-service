package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/m1tka051209/calculator-service/calculator"
	"github.com/m1tka051209/calculator-service/config"
	"github.com/m1tka051209/calculator-service/db"
	"github.com/m1tka051209/calculator-service/models"
	"github.com/m1tka051209/calculator-service/task_manager"
)

func main() {
	cfg := config.Load()

	// Настройка логгера
	logFile, err := os.OpenFile(cfg.WorkerLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// Инициализация репозитория
	repo, err := db.NewSQLiteRepository(cfg.WorkerDBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer repo.Close()

	// Инициализация менеджера задач
	tm := task_manager.NewTaskManager(repo)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обработка сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запуск воркеров
	for i := 0; i < cfg.WorkerPoolSize; i++ {
		go worker(ctx, tm, i)
	}

	log.Println("Worker started successfully")
	log.Printf("Using DB path: %s", cfg.WorkerDBPath)
	log.Printf("Using log path: %s", cfg.WorkerLogPath)
	log.Printf("Worker pool size: %d", cfg.WorkerPoolSize)

	// Ожидание сигнала завершения
	<-sigChan
	log.Println("Received termination signal, shutting down...")
	cancel()
	time.Sleep(1 * time.Second) // Даем воркерам время завершиться
	log.Println("Worker stopped gracefully")
}

func worker(ctx context.Context, tm *task_manager.TaskManager, workerID int) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d shutting down", workerID)
			return
		case <-ticker.C:
			processTask(ctx, tm, workerID)
		}
	}
}

func processTask(ctx context.Context, tm *task_manager.TaskManager, workerID int) {
	// Получаем следующую задачу
	task, err := tm.GetNextTask()
	if err != nil {
		log.Printf("Worker %d error getting task: %v", workerID, err)
		time.Sleep(2 * time.Second)
		return
	}

	if task == nil {
		// Нет задач для обработки
		return
	}

	log.Printf("Worker %d started processing task %s", workerID, task.ID)

	// Обновляем статус задачи на "processing"
	if err := tm.UpdateTaskStatus(ctx, task.ID, "processing"); err != nil {
		log.Printf("Worker %d error updating task status: %v", workerID, err)
		return
	}

	// Создаем структуру для калькулятора
	calcTask := &models.Task{
		ID:            task.ID,
		Arg1:          task.Arg1,
		Arg2:          task.Arg2,
		Operation:     task.Operation,
		OperationTime: task.OperationTime,
	}

	// Вычисляем результат
	result := calculator.Calculate(calcTask)

	// Сохраняем результат
	if err := tm.SaveTaskResult(task.ID, result); err != nil {
		log.Printf("Worker %d error saving task result: %v", workerID, err)
		return
	}

	log.Printf("Worker %d completed task %s with result: %f", workerID, task.ID, result)
}
