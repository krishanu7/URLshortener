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
			"INSERT INTO urls (original_url, short_code) VALUES ($1, $2) RETURNING id, created_at, updated_at, access_count, is_active", url.OriginalURL, url.ShortCode,
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

func (r *URLRepository) GetByShortCode(shortCode string) (*models.URL, error) {
	url := &models.URL{}
	err := r.db.QueryRow(
		"SELECT id, original_url, short_code, created_at, updated_at, access_count, is_active FROM urls WHERE short_code = $1", shortCode, 
	).Scan(&url.ID, &url.OriginalURL, &url.ShortCode, &url.CreatedAt, &url.UpdatedAt, &url.AccessCount, &url.IsActive)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return url, nil
}

func (r *URLRepository) Update(shortCode string, originalURL string) (*models.URL, error) {
    url := &models.URL{}
    err := r.db.QueryRow(
        "UPDATE urls SET original_url = $1, updated_at = CURRENT_TIMESTAMP WHERE short_code = $2, is_active = $3 RETURNING id, original_url, short_code, created_at, updated_at, access_count, is_active",
        originalURL, shortCode,
    ).Scan(&url.ID, &url.OriginalURL, &url.ShortCode, &url.CreatedAt, &url.UpdatedAt, &url.AccessCount, &url.IsActive)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    return url, nil
}

func (r *URLRepository) Delete(shortCode string) error {
    result, err := r.db.Exec("DELETE FROM urls WHERE short_code = $1", shortCode)
    if err != nil {
        return err
    }
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rowsAffected == 0 {
        return sql.ErrNoRows
    }
    return nil
}

func (r *URLRepository) IncrementAccessCount(shortCode string) error {
	result, err := r.db.Exec("UPDATE urls SET access_count = access_count + 1 WHERE short_code = $1", shortCode)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}