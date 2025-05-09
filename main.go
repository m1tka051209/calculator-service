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

	// Инициализация репозитория
	repo, err := db.NewSQLiteRepository(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}
	defer repo.Close()

	// Запуск gRPC сервера
	go func() {
		if err := server.StartGRPCServer(cfg.GRPCPort, repo); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Запуск HTTP сервера
	http.Handle("/", api.StartHTTPGateway())
	log.Println("HTTP server started on :8080")
	go http.ListenAndServe(":8080", nil)

	// Запуск воркеров
	worker.RunWorker(repo, cfg.WorkerPoolSize)
}