package main

import (
    "database/sql"
    "log"
    "net/http"
    "urlshortener/internal/config"
    "urlshortener/internal/handlers"
    "urlshortener/internal/repository"

    "github.com/go-redis/redis/v8"
    "github.com/gorilla/mux"
    _ "github.com/lib/pq"
)

func main() {
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    db, err := sql.Open("postgres", cfg.DatabaseURL)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Fatalf("Database ping failed: %v", err)
    }

    redisClient := redis.NewClient(&redis.Options{
        Addr: cfg.RedisURL,
    })
    if _, err := redisClient.Ping(redisClient.Context()).Result(); err != nil {
        log.Fatalf("Failed to connect to Redis: %v", err)
    }
    defer redisClient.Close()

    urlRepo := repository.NewURLRepository(db, redisClient)
    urlHandler := handlers.NewURLHandler(urlRepo)

    r := mux.NewRouter()
    r.HandleFunc("/shorten", urlHandler.Shorten).Methods("POST")
    r.HandleFunc("/shorten/{code}", urlHandler.Get).Methods("GET")
    r.HandleFunc("/shorten/{code}", urlHandler.Update).Methods("PUT")
    r.HandleFunc("/shorten/{code}", urlHandler.Delete).Methods("DELETE")
    r.HandleFunc("/shorten/{code}/stats", urlHandler.Stats).Methods("GET")
    r.HandleFunc("/{code}", urlHandler.Redirect).Methods("GET")

    log.Printf("Starting server on :%s", cfg.Port)
    if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}