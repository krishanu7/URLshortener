package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"urlshortener/internal/models"
	"urlshortener/internal/repository"
)


type URLHandler struct {
	repo *repository.URLRepository
}

func NewURLHandler(repo *repository.URLRepository) *URLHandler {
	return &URLHandler{
		repo: repo,
	}
}

func (h *URLHandler) Shorten(w http.ResponseWriter, r *http.Request) {
	var req models.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
        return
    }
	// Validate URL
	if _, err := url.ParseRequestURI(req.URL); err != nil {
		http.Error(w, `{"error": "Invalid URL"}`, http.StatusBadRequest)
        return
	}
	urlModel := &models.URL{OriginalURL: req.URL}

	if err := h.repo.Create(urlModel); err != nil {
		http.Error(w,`{"error": "Failed to create short URL"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(urlModel)
}