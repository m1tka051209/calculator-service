package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPPort        string
	DBPath          string
	JWTSecret       string
	TokenExpiration time.Duration
	WorkerPoolSize  int
	WorkerTimeout   time.Duration
}

func Load() *Config {
	return &Config{
		HTTPPort:        getEnv("HTTP_PORT", "8080"),
		DBPath:          getEnv("DB_PATH", "data.db"),
		JWTSecret:       getEnv("JWT_SECRET", "default-secret"),
		TokenExpiration: getEnvAsDuration("TOKEN_EXPIRATION", 24*time.Hour),
		WorkerPoolSize:  getEnvAsInt("WORKER_POOL_SIZE", 5),
		WorkerTimeout:   getEnvAsDuration("WORKER_TIMEOUT", 5*time.Second),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if num, err := strconv.Atoi(value); err == nil {
			return num
		}
	}
	return defaultValue
}