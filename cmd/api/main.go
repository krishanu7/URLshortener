package main

import (
	"database/sql"
	"log"
	"net/http"
	"urlshortener/internal/config"
	"urlshortener/internal/handlers"
	"urlshortener/internal/middleware"
	"urlshortener/internal/repository"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

func main() {
    logger := logrus.New()
    logger.SetFormatter(&logrus.JSONFormatter{})
    logger.SetLevel(logrus.InfoLevel)

    cfg, err := config.LoadConfig()
    if err != nil {
        logger.WithError(err).Fatal("Failed to load config")
    }

    db, err := sql.Open("postgres", cfg.DatabaseURL)
    if err != nil {
         logger.WithError(err).Fatal("Failed to connect to database")
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        logger.WithError(err).Fatal("Database ping failed")
    }

    redisClient := redis.NewClient(&redis.Options{
        Addr: cfg.RedisURL,
    })
    if _, err := redisClient.Ping(redisClient.Context()).Result(); err != nil {
        logger.WithError(err).Fatal("Failed to connect to Redis")
    }
    defer redisClient.Close()

    urlRepo := repository.NewURLRepository(db, redisClient, logger)
    urlHandler := handlers.NewURLHandler(urlRepo, logger)

    r := mux.NewRouter()
    // Middleware
    rateLimiter := middleware.NewRateLimiter(rate.Limit(10.0/60.0), 10)
    r.Use(middleware.LoggingMiddleware(logger))
    r.Use(rateLimiter.Middleware)

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