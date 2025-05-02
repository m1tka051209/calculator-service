package main

import (
	"log"
	"net/http"

	"github.com/m1tka051209/calculator-service/api"
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

	tm := task_manager.NewTaskManager(repo)
	handlers := api.NewHandlers(repo, cfg.JWTSecret, cfg.TokenExpiration)

	http.HandleFunc("/api/v1/register", handlers.RegisterHandler)
	http.HandleFunc("/api/v1/login", handlers.LoginHandler)
	http.Handle("/api/v1/calculate", api.AuthMiddleware(handlers, http.HandlerFunc(handlers.CalculateHandler)))

	log.Printf("Server started on :%s", cfg.HTTPPort)
	log.Fatal(http.ListenAndServe(":"+cfg.HTTPPort, nil))
}