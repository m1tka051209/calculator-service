package worker

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/m1tka051209/calculator-service/db"
	"github.com/m1tka051209/calculator-service/server"
	"github.com/m1tka051209/calculator-service/task_manager"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// CalculatorClient определяет упрощенный интерфейс клиента
type CalculatorClient interface {
	Calculate(ctx context.Context, req *server.CalculationRequest) (*server.CalculationResponse, error)
}

// grpcClientWrapper оборачивает gRPC клиент для соответствия интерфейсу
type grpcClientWrapper struct {
	client server.CalculatorClient
}

func (w *grpcClientWrapper) Calculate(ctx context.Context, req *server.CalculationRequest) (*server.CalculationResponse, error) {
	return w.client.Calculate(ctx, req)
}

func NewCalculatorClient(conn *grpc.ClientConn) CalculatorClient {
	return &grpcClientWrapper{
		client: server.NewCalculatorClient(conn),
	}
}

func RunWorker(repo db.Repository, workerCount int) {
	tm := task_manager.NewTaskManager(repo)

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := NewCalculatorClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < workerCount; i++ {
		go processTasks(ctx, tm, client, i)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutting down workers...")
	cancel()
	time.Sleep(1 * time.Second)
}

func processTasks(ctx context.Context, tm task_manager.TaskManagerInterface, client CalculatorClient, workerID int) {
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

			resp, err := client.Calculate(ctx, &server.CalculationRequest{
				Arg1:      task.Arg1,
				Arg2:      task.Arg2,
				Operation: task.Operation,
			})

			if err != nil {
				log.Printf("Worker %d calculation error: %v", workerID, err)
				tm.UpdateTaskStatus(ctx, task.ID, "failed")
				continue
			}

			if err := tm.UpdateTaskResult(ctx, task.ID, resp.Result); err != nil {
				log.Printf("Worker %d error saving result: %v", workerID, err)
			} else {
				log.Printf("Worker %d successfully processed task %s", workerID, task.ID)
			}
		}
	}
}