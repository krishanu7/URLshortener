package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
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

func (h *URLHandler) Get(w http.ResponseWriter, r *http.Request) {
	shortCode := strings.TrimPrefix(r.URL.Path, "/shorten/")
	url, err := h.repo.GetByShortCode(shortCode)

	if err != nil {
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	if url == nil {
		http.Error(w, `{"error": "Short URL not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(url)
}

func (h *URLHandler) Update(w http.ResponseWriter, r *http.Request) {
	shortCode := strings.TrimPrefix(r.URL.Path, "/shorten/")
	var req models.ShortenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
        return
	}

	if _, err := url.ParseRequestURI(req.URL); err != nil {
        http.Error(w, `{"error": "Invalid URL"}`, http.StatusBadRequest)
        return
    }
	url, err := h.repo.Update(shortCode, req.URL);

	if err != nil {
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
        return
	}

	if url == nil {
		http.Error(w, `{"error": "Short URL not found"}`, http.StatusNotFound)
        return
	}

	w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(url)
}

func (h *URLHandler) Delete(w http.ResponseWriter, r *http.Request) {
    shortCode := strings.TrimPrefix(r.URL.Path, "/shorten/")
    err := h.repo.Delete(shortCode)
    if err == sql.ErrNoRows {
        http.Error(w, `{"error": "Short URL not found"}`, http.StatusNotFound)
        return
    }
    if err != nil {
        http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

func (h *URLHandler) Stats(w http.ResponseWriter, r *http.Request) {
	shortCode := strings.TrimPrefix(r.URL.Path, "/shorten/")
	shortCode = strings.TrimSuffix(shortCode, "/stats")
	url, err := h.repo.GetByShortCode(shortCode);
	
	if err != nil {
        http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
        return
    }
    if url == nil {
        http.Error(w, `{"error": "Short URL not found"}`, http.StatusNotFound)
        return
    }
	w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(url)
}

func (h *URLHandler) Redirect(w http.ResponseWriter, r *http.Request) {
    shortCode := strings.TrimPrefix(r.URL.Path, "/")
    url, err := h.repo.GetByShortCode(shortCode)
    if err != nil {
        http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
        return
    }
    if url == nil {
        http.Error(w, `{"error": "Short URL not found"}`, http.StatusNotFound)
        return
    }

    if err := h.repo.IncrementAccessCount(shortCode); err != nil {
        http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, url.OriginalURL, http.StatusMovedPermanently)
}