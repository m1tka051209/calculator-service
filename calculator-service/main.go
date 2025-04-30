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
    handlers := api.NewHandlers(tm, repo)

    // Public API
    http.HandleFunc("/api/v1/register", handlers.RegisterHandler)
    http.HandleFunc("/api/v1/login", handlers.LoginHandler)
	http.Handle("/api/v1/calculate", api.AuthMiddleware(repo, cfg.JWTSecret)(http.HandlerFunc(handlers.CalculateHandler)))
    // Internal API for workers
    http.HandleFunc("/internal/task", handlers.TaskHandler)
    http.HandleFunc("/internal/result", handlers.ResultHandler)

    log.Printf("🚀 Server started on :%s", cfg.HTTPPort)
    log.Fatal(http.ListenAndServe(":"+cfg.HTTPPort, nil))
}