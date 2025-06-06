package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)


type Config struct {
	DatabaseURL string
	Port string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, useing default environment variables")
	}
	config := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port: os.Getenv("PORT"),
	}

	if config.DatabaseURL == "" {
		config.DatabaseURL = "postgres://urlshortener:mysecretpassword@localhost:5432/urlshortener?sslmode=disable"
	}

	if config.Port == "" {
		config.Port = "8080"
	}
	return config, nil
}