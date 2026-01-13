package utils

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// SendJSONError sends a JSON error response
func SendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

// SendJSONResponse sends a JSON success response
func SendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// IsValidURL validates if a string is a valid HTTP/HTTPS URL
func IsValidURL(urlStr string) bool {
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return false
	}
	u, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// IsValidCode validates if a custom code contains only allowed characters
func IsValidCode(code string) bool {
	if len(code) < 3 || len(code) > 20 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, code)
	return matched
}

// GenerateRandomCode generates a random alphanumeric code of specified length
func GenerateRandomCode(length int) (string, error) {
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
