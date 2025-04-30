package config

import (
    "os"
    "time"
)

type Config struct {
    HTTPPort         string
    DBPath          string
    JWTSecret       string
    TokenExpiration time.Duration
}

func Load() *Config {
    return &Config{
        HTTPPort:         getEnv("HTTP_PORT", "8080"),
        DBPath:          getEnv("DB_PATH", "data.db"),
        JWTSecret:       getEnv("JWT_SECRET", "default-secret-key"),
        TokenExpiration: getEnvAsDuration("TOKEN_EXPIRATION", 24*time.Hour),
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