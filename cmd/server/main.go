package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"

	"url_shortener/internal/config"
	"url_shortener/internal/storage"

	"github.com/joho/godotenv"
)

type shortenRequest struct {
	URL        string `json:"url"`
	CustomCode string `json:"custom_code,omitempty"`
	TTL        int    `json:"ttl,omitempty"` // TTL in seconds, 0 means no expiration
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
	Code     string `json:"code"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// sendJSONError sends a JSON error response
func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse{Error: message})
}

// sendJSONResponse sends a JSON success response
func sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// isValidURL validates if a string is a valid HTTP/HTTPS URL
func isValidURL(urlStr string) bool {
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return false
	}
	u, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// isValidCode validates if a custom code contains only allowed characters
func isValidCode(code string) bool {
	if len(code) < 3 || len(code) > 20 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, code)
	return matched
}

// generateRandomCode generates a random alphanumeric code of specified length
func generateRandomCode(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}

func main() {
	r := chi.NewRouter()
	r.Use(config.CorsConfig())
	godotenv.Load()

	redis := storage.NewRedis()

	if err := redis.Ping(context.Background()); err != nil {
		log.Fatal("Redis not connected:", err)
	}

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var req shortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendJSONError(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate URL
		if req.URL == "" {
			sendJSONError(w, "url is required", http.StatusBadRequest)
			return
		}

		if !isValidURL(req.URL) {
			sendJSONError(w, "invalid URL format, must start with http:// or https://", http.StatusBadRequest)
			return
		}

		// Validate or generate code
		code := req.CustomCode

		if code == "" {
			// Auto-generate code if not provided
			var err error
			for attempts := 0; attempts < 5; attempts++ {
				code, err = generateRandomCode(6)
				if err != nil {
					continue
				}
				// Check if code is available
				_, err := redis.GetKeyValue(r.Context(), "short:"+code)
				if err != nil {
					// Code is available (key not found)
					break
				}
			}
			if err != nil {
				sendJSONError(w, "failed to generate unique code", http.StatusInternalServerError)
				return
			}
		} else {
			// Validate custom code
			if !isValidCode(code) {
				sendJSONError(w, "invalid custom code, must be 3-20 characters and contain only letters, numbers, hyphens, and underscores", http.StatusBadRequest)
				return
			}
		}

		// Store with or without TTL
		var ok bool
		var err error
		if req.TTL > 0 {
			ok, err = redis.SetKeyValueWithTTL(r.Context(), "shorten.link/"+code, req.URL, req.TTL)
		} else {
			ok, err = redis.SetKeyValue(r.Context(), "shorten.link/"+code, req.URL)
		}

		if err != nil {
			sendJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if !ok {
			sendJSONError(w, "short URL already taken", http.StatusConflict)
			return
		}

		shortURL := ""

		if os.Getenv("SHORT_API") != "" {
			shortURL = os.Getenv("SHORT_API") + code
		}

		sendJSONResponse(w, shortenResponse{
			ShortURL: shortURL,
			Code:     code,
		}, http.StatusCreated)
	})

	r.Get("/{code}", func(w http.ResponseWriter, r *http.Request) {
		code := chi.URLParam(r, "code")

		url, err := redis.GetKeyValue(r.Context(), "shorten.link/"+code)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, url, http.StatusFound)
	})

	log.Println("Listening on port 8080")
	http.ListenAndServe(":8080", r)
}
