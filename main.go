package main

import (
	"log"
	"net/http"
	
	"github.com/m1tka051209/calculator-service/api"
	"github.com/m1tka051209/calculator-service/config"
	"github.com/m1tka051209/calculator-service/db"
	"github.com/m1tka051209/calculator-service/server"
	"github.com/m1tka051209/calculator-service/worker"
)

func main() {
	cfg := config.Load()

	repo, err := db.NewSQLiteRepository(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer repo.Close()

	// Запуск gRPC сервера
	go func() {
		if err := server.StartGRPCServer(cfg.GRPCPort, repo); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Запуск HTTP сервера
	httpHandler := api.StartHTTPGateway()
	go func() {
		log.Println("HTTP server starting on :8080")
		if err := http.ListenAndServe(":8080", httpHandler); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Запуск воркеров
	worker.RunWorker(repo, cfg.WorkerPoolSize)
}