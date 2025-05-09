package main

import (
	"log"

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

	// Запуск воркеров
	worker.RunWorker(repo, cfg.WorkerPoolSize)
}