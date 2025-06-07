package models

import "time"

type URL struct {
	ID          int       `json:"id"`
	OriginalURL string    `json:"url"`
	ShortCode   string    `json:"shortCode"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	AccessCount int       `json:"accessCount"`
	IsActive    bool      `json:"isActive"`
}

type ShortenRequest struct {
	URL string `json:"url"`
}
