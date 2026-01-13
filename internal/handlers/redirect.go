package handlers

import (
	"net/http"

	"url_shortener/internal/storage"

	"github.com/go-chi/chi/v5"
)

type RedirectHandler struct {
	Redis *storage.RedisClient
}

func (h *RedirectHandler) RedirectToURL(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	if code == "" {
		http.NotFound(w, r)
		return
	}

	url, err := h.Redis.GetKeyValue(r.Context(), "localhost:8080/"+code)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}
