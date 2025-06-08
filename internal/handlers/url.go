package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"urlshortener/internal/models"
	"urlshortener/internal/repository"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type URLHandler struct {
    repo *repository.URLRepository
    logger *logrus.Logger
}

func NewURLHandler(repo *repository.URLRepository, logger *logrus.Logger) *URLHandler {
    return &URLHandler{repo: repo, logger: logger }
}

func (h *URLHandler) Shorten(w http.ResponseWriter, r *http.Request) {
    var req models.ShortenRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.WithError(err).Error("Invalid request body for shorten")
        http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
        return
    }

    if _, err := url.ParseRequestURI(req.URL); err != nil {
        h.logger.WithError(err).WithField("url", req.URL).Error("Invalid URL")
        http.Error(w, `{"error": "Invalid URL"}`, http.StatusBadRequest)
        return
    }

    urlModel := &models.URL{OriginalURL: req.URL}
    if err := h.repo.Create(urlModel); err != nil {
        h.logger.WithError(err).Error("Failed to create short URL")
        http.Error(w, `{"error": "Failed to create short URL"}`, http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(urlModel)
    h.logger.WithFields(logrus.Fields{
        "short_code":   urlModel.ShortCode,
        "original_url": urlModel.OriginalURL,
    }).Info("Short URL created")
}

func (h *URLHandler) Get(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    shortCode := vars["code"]
    url, err := h.repo.GetByShortCode(shortCode)
    if err != nil {
        h.logger.WithError(err).WithField("short_code", shortCode).Error("Failed to fetch URL")
        http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
        return
    }
    if url == nil {
        h.logger.WithField("short_code", shortCode).Warn("Short URL not found")
        http.Error(w, `{"error": "Short URL not found"}`, http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(url)
    h.logger.WithField("short_code", shortCode).Info("Retrieved short URL")
}

func (h *URLHandler) Update(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    shortCode := vars["code"]
    var req models.ShortenRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.WithError(err).Error("Invalid request body for update")
        http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
        return
    }

    if _, err := url.ParseRequestURI(req.URL); err != nil {
        h.logger.WithError(err).WithField("url", req.URL).Error("Invalid URL")
        http.Error(w, `{"error": "Invalid URL"}`, http.StatusBadRequest)
        return
    }

    url, err := h.repo.Update(shortCode, req.URL)
    if err != nil {
        h.logger.WithError(err).WithField("short_code", shortCode).Error("Failed to update URL")
        http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
        return
    }
    if url == nil {
        h.logger.WithField("short_code", shortCode).Warn("Short URL not found for update")
        http.Error(w, `{"error": "Short URL not found"}`, http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(url)
    h.logger.WithFields(logrus.Fields{
        "short_code":   shortCode,
        "original_url": req.URL,
    }).Info("Updated short URL")
}

func (h *URLHandler) Delete(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    shortCode := vars["code"]
    err := h.repo.Delete(shortCode)
    if err == sql.ErrNoRows {
        h.logger.WithField("short_code", shortCode).Warn("Short URL not found for delete")
        http.Error(w, `{"error": "Short URL not found"}`, http.StatusNotFound)
        return
    }
    if err != nil {
        h.logger.WithError(err).WithField("short_code", shortCode).Error("Failed to delete URL")
        http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
    h.logger.WithField("short_code", shortCode).Info("Deleted short URL")
}

func (h *URLHandler) Stats(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    shortCode := vars["code"]
    url, err := h.repo.GetByShortCode(shortCode)
    if err != nil {
        h.logger.WithError(err).WithField("short_code", shortCode).Error("Failed to fetch URL stats")
        http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
        return
    }
    if url == nil {
        h.logger.WithField("short_code", shortCode).Warn("Short URL not found for stats")
        http.Error(w, `{"error": "Short URL not found"}`, http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(url)
    h.logger.WithField("short_code", shortCode).Info("Retrieved URL stats")
}

func (h *URLHandler) Redirect(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    shortCode := vars["code"]
    url, err := h.repo.GetByShortCode(shortCode)
    if err != nil {
        h.logger.WithError(err).WithField("short_code", shortCode).Error("Failed to fetch URL for redirect")
        http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
        return
    }
    if url == nil {
        h.logger.WithField("short_code", shortCode).Warn("Short URL not found for redirect")
        http.Error(w, `{"error": "Short URL not found"}`, http.StatusNotFound)
        return
    }

    if err := h.repo.IncrementAccessCount(shortCode); err != nil {
        h.logger.WithError(err).WithField("short_code", shortCode).Error("Failed to increment access count")
        http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, url.OriginalURL, http.StatusMovedPermanently)
    h.logger.WithFields(logrus.Fields{
        "short_code":   shortCode,
        "original_url": url.OriginalURL,
    }).Info("Redirected to original URL")
}