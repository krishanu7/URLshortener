package main

import (
	"database/sql"
	"log"
	"net/http"
	"urlshortener/internal/config"
	"urlshortener/internal/handlers"
	"urlshortener/internal/repository"

	_ "github.com/lib/pq"
)



func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	db, err := sql.Open("postgres", cfg.DatabaseURL);
	if err != nil {
		 log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	urlRepo := repository.NewURLRepository(db)
	urlHandler := handlers.NewURLHandler(urlRepo)

	http.HandleFunc("/shorten", urlHandler.Shorten)
	
	log.Printf("Starting server on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}