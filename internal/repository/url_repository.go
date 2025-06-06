package repository

import (
	"database/sql"
	"urlshortener/internal/models"
	"urlshortener/internal/util"
)



type URLRepository struct {
	db *sql.DB
}

func NewURLRepository(db *sql.DB) *URLRepository {
	return &URLRepository{
		db: db,
	}
}

func (r *URLRepository) Create(url *models.URL) error {
	for {
		shortCode, err := util.GenerateShortCode(7)
		if err != nil {
			return err
		}
		url.ShortCode = shortCode
		err = r.db.QueryRow(
			"INSERT INTO urls (original_url, short_code) VALUES ($1, $2) RETURING id, created_at, updated_at, access_count, is_active", url.OriginalURL, url.ShortCode,
		).Scan(&url.ID, &url.CreatedAt, &url.UpdatedAt, &url.AccessCount, &url.IsActive)

		if err == nil {
			return nil
		}
		if err.Error() == `pq: duplicate key value violates unique constraint "urls_short_code_key"` {
            continue
        }
        return err
	}
}