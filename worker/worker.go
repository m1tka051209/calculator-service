package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/m1tka051209/calculator-service/db"
	"github.com/m1tka051209/calculator-service/models"
	"github.com/m1tka051209/calculator-service/task_manager"
	"google.golang.org/grpc"
)

func RunWorker(repo db.Repository, workerCount int) {
	tm := task_manager.NewTaskManager(repo)

	// Подключение к gRPC серверу
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запуск воркеров
	for i := 0; i < workerCount; i++ {
		go processTasks(ctx, tm, i)
	}

	// Обработка сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutting down workers...")
	cancel()
	time.Sleep(1 * time.Second) // Даем время для graceful shutdown
}

func processTasks(ctx context.Context, tm *task_manager.TaskManager, workerID int) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d shutting down", workerID)
			return
		case <-ticker.C:
			task, err := tm.GetNextTask()
			if err != nil {
				log.Printf("Worker %d error getting task: %v", workerID, err)
				continue
			}

			if task == nil {
				continue
			}

			// Обновляем статус задачи
			if err := tm.UpdateTaskStatus(ctx, task.ID, "processing"); err != nil {
				log.Printf("Worker %d error updating status: %v", workerID, err)
				continue
			}

			// Вычисляем результат
			result, err := calculate(task)
			if err != nil {
				log.Printf("Worker %d calculation error: %v", workerID, err)
				if err := tm.UpdateTaskStatus(ctx, task.ID, "failed"); err != nil {
					log.Printf("Worker %d error marking task as failed: %v", workerID, err)
				}
				continue
			}

			// Сохраняем результат
			if err := tm.UpdateTaskResult(ctx, task.ID, result); err != nil {
				log.Printf("Worker %d error saving result: %v", workerID, err)
			} else {
				log.Printf("Worker %d successfully processed task %s", workerID, task.ID)
			}
		}
	}
}

func calculate(task *models.Task) (float64, error) {
	// Имитация времени выполнения операции
	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)

	switch task.Operation {
	case "+":
		return task.Arg1 + task.Arg2, nil
	case "-":
		return task.Arg1 - task.Arg2, nil
	case "*":
		return task.Arg1 * task.Arg2, nil
	case "/":
		if task.Arg2 == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return task.Arg1 / task.Arg2, nil
	default:
		return 0, fmt.Errorf("unknown operation: %s", task.Operation)
	}
}