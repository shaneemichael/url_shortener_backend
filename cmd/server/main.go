package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"url_shortener/internal/config"
	"url_shortener/internal/handlers"
	"url_shortener/internal/middleware"
	"url_shortener/internal/storage"
)

func main() {
	// Load environment variables
	godotenv.Load()

	// Initialize Redis
	redis := storage.NewRedis()
	if err := redis.Ping(context.Background()); err != nil {
		log.Fatal("Redis not connected:", err)
	}
	log.Println("Redis connected successfully")

	// Initialize handlers
	shortenHandler := &handlers.ShortenHandler{Redis: redis}
	redirectHandler := &handlers.RedirectHandler{Redis: redis}

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(config.CorsConfig())

	// Routes
	r.Post("/", shortenHandler.CreateShortURL)
	r.Get("/{code}", redirectHandler.RedirectToURL)

	// Start server
	log.Println("Server starting on port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
