package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"url_shortener/internal/storage"
	"url_shortener/internal/utils"
)

type ShortenHandler struct {
	Redis *storage.RedisClient
}

type shortenRequest struct {
	URL        string `json:"url"`
	CustomCode string `json:"custom_code,omitempty"`
	TTL        int    `json:"ttl,omitempty"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
	Code     string `json:"code"`
}

func (h *ShortenHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendJSONError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate URL
	if req.URL == "" {
		utils.SendJSONError(w, "url is required", http.StatusBadRequest)
		return
	}

	if !utils.IsValidURL(req.URL) {
		utils.SendJSONError(w, "invalid URL format, must start with http:// or https://", http.StatusBadRequest)
		return
	}

	// Validate or generate code
	code := req.CustomCode
	if code == "" {
		code = h.generateUniqueCode(r.Context())
		if code == "" {
			utils.SendJSONError(w, "failed to generate unique code", http.StatusInternalServerError)
			return
		}
	} else {
		if !utils.IsValidCode(code) {
			utils.SendJSONError(w, "invalid custom code, must be 3-20 characters and contain only letters, numbers, hyphens, and underscores", http.StatusBadRequest)
			return
		}
	}

	// Store with or without TTL
	key := "short:" + code
	var ok bool
	var err error

	if req.TTL > 0 {
		ok, err = h.Redis.SetKeyValueWithTTL(r.Context(), key, req.URL, req.TTL)
	} else {
		ok, err = h.Redis.SetKeyValue(r.Context(), key, req.URL)
	}

	if err != nil {
		utils.SendJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if !ok {
		utils.SendJSONError(w, "short URL already taken", http.StatusConflict)
		return
	}

	// Build short URL
	shortURL := h.buildShortURL(code)

	utils.SendJSONResponse(w, shortenResponse{
		ShortURL: shortURL,
		Code:     code,
	}, http.StatusCreated)
}

func (h *ShortenHandler) generateUniqueCode(ctx context.Context) string {
	for attempts := 0; attempts < 5; attempts++ {
		code, err := utils.GenerateRandomCode(6)
		if err != nil {
			continue
		}

		// Check if code is available
		_, err = h.Redis.GetKeyValue(ctx, "short:"+code)
		if err != nil {
			// Code is available (key not found)
			return code
		}
	}
	return ""
}

func (h *ShortenHandler) buildShortURL(code string) string {
	baseURL := os.Getenv("SHORT_API")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return baseURL + "/" + code
}
