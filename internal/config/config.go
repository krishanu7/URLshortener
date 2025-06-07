package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
	RedisURL    string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, useing default environment variables")
	}
	config := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
		RedisURL:    os.Getenv("REDIS_URL"),
	}

	if config.DatabaseURL == "" {
		config.DatabaseURL = "postgres://urlshortener:mysecretpassword@localhost:5432/urlshortener?sslmode=disable"
	}

	if config.Port == "" {
		config.Port = "8080"
	}

	if config.RedisURL == "" {
		config.RedisURL = "redis://localhost:6379"
	}
	return config, nil
}
