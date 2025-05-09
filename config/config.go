package config

import (
	"os"
	"strconv"
)

type Config struct {
	GRPCPort       string
	DBPath         string
	WorkerPoolSize int
}

func Load() *Config {
	return &Config{
		GRPCPort:       getEnv("GRPC_PORT", "50051"),
		DBPath:         getEnv("DB_PATH", "data.db"),
		WorkerPoolSize: getEnvAsInt("WORKER_POOL_SIZE", 3),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
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