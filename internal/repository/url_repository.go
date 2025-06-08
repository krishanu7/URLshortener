package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
	"urlshortener/internal/models"
	"urlshortener/internal/util"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type URLRepository struct {
	db     *sql.DB
	redis  *redis.Client
	ctx    context.Context
	logger *logrus.Logger
}

func NewURLRepository(db *sql.DB, redis *redis.Client, logger *logrus.Logger) *URLRepository {
	repo := &URLRepository{
		db:     db,
		redis:  redis,
		ctx:    context.Background(),
		logger: logger,
	}
	go repo.startAccessCountSync(30 * time.Second)
	return repo
}

func (r *URLRepository) Create(url *models.URL) error {
	for {
		shortCode, err := util.GenerateShortCode(7)
		if err != nil {
			r.logger.WithError(err).Error("Failed to generate short code")
			return err
		}
		url.ShortCode = shortCode
		err = r.db.QueryRow(
			"INSERT INTO urls (original_url, short_code) VALUES ($1, $2) RETURNING id, created_at, updated_at, access_count, is_active", url.OriginalURL, url.ShortCode,
		).Scan(&url.ID, &url.CreatedAt, &url.UpdatedAt, &url.AccessCount, &url.IsActive)

		if err == nil {
			urlData, _ := json.Marshal(url)
			r.redis.Set(r.ctx, "url:"+url.ShortCode, urlData, 24*time.Hour)
			r.logger.WithFields(logrus.Fields{
				"short_code":   url.ShortCode,
				"original_url": url.OriginalURL,
			}).Info("Created short URL")
			return nil
		}
		if err.Error() == `pq: duplicate key value violates unique constraint "urls_short_code_key"` {
			r.logger.WithField("short_code", shortCode).Warn("Short code collision, retrying")
			continue
		}
		r.logger.WithError(err).Error("Failed to create short URL in database")
		return err
	}
}

func (r *URLRepository) GetByShortCode(shortCode string) (*models.URL, error) {
	// Check Redis first
	cached, err := r.redis.Get(r.ctx, "url:"+shortCode).Result()

	if err == nil {
		var url models.URL
		if json.Unmarshal([]byte(cached), &url) == nil {
			r.logger.WithField("short_code", shortCode).Info("Retrieved URL from Redis")
			return &url, nil
		}
	}
	if err != redis.Nil {
		r.logger.WithError(err).Warn("Redis error while fetching URL")
	}
	// Fallback to Postgresql
	url := &models.URL{}

	err = r.db.QueryRow(
		"SELECT id, original_url, short_code, created_at, updated_at, access_count FROM urls WHERE short_code = $1", shortCode,
	).Scan(&url.ID, &url.OriginalURL, &url.ShortCode, &url.CreatedAt, &url.UpdatedAt, &url.AccessCount)
	if err == sql.ErrNoRows {
		r.logger.WithField("short_code", shortCode).Warn("Short URL not found in database")
		return nil, nil
	}
	if err != nil {
		r.logger.WithError(err).Error("Failed to fetch URL from database")
		return nil, err
	}
	// Cache the result
	urlData, _ := json.Marshal(url)
	r.redis.Set(r.ctx, "url:"+shortCode, urlData, 24*time.Hour)
	r.logger.WithField("short_code", shortCode).Info("Retrieved URL from database and cached in Redis")
	return url, nil
}

func (r *URLRepository) Update(shortCode string, originalURL string) (*models.URL, error) {
	url := &models.URL{}
	err := r.db.QueryRow(
		"UPDATE urls SET original_url = $1, updated_at = CURRENT_TIMESTAMP WHERE short_code = $2, is_active = $3 RETURNING id, original_url, short_code, created_at, updated_at, access_count, is_active",
		originalURL, shortCode,
	).Scan(&url.ID, &url.OriginalURL, &url.ShortCode, &url.CreatedAt, &url.UpdatedAt, &url.AccessCount, &url.IsActive)
	if err == sql.ErrNoRows {
		r.logger.WithField("short_code", shortCode).Warn("Short URL not found for update")
		return nil, nil
	}
	if err != nil {
		r.logger.WithError(err).Error("Failed to update URL in database")
		return nil, err
	}
	// Update Cache
	urlData, _ := json.Marshal(url)
	r.redis.Set(r.ctx, "url:"+shortCode, urlData, 24*time.Hour)
	r.logger.WithFields(logrus.Fields{
		"short_code":   shortCode,
		"original_url": originalURL,
	}).Info("Updated short URL")
	return url, nil
}

func (r *URLRepository) Delete(shortCode string) error {
	result, err := r.db.Exec("DELETE FROM urls WHERE short_code = $1", shortCode)
	if err != nil {
		r.logger.WithError(err).Error("Failed to delete URL from database")
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.WithError(err).Error("Failed to check rows affected for delete")
		return err
	}
	if rowsAffected == 0 {
		r.logger.WithField("short_code", shortCode).Warn("Short URL not found for delete")
		return sql.ErrNoRows
	}
	// Invalidate Cache
	r.redis.Del(r.ctx, "url:"+shortCode)
	r.redis.Del(r.ctx, "access_count:"+shortCode)
	r.logger.WithField("short_code", shortCode).Info("Deleted short URL")
	return nil
}

func (r *URLRepository) IncrementAccessCount(shortCode string) error {
	// Increment in Redis
	_, err := r.redis.Incr(r.ctx, "access_count:"+shortCode).Result()
	if err != nil {
		r.logger.WithError(err).Error("Failed to increment access count in Redis")
		return err
	}
	r.logger.WithField("short_code", shortCode).Info("Incremented access count in Redis")
	return nil
}

func (r *URLRepository) startAccessCountSync(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		r.syncAccessCounts()
	}
}

func (r *URLRepository) syncAccessCounts() {
	keys, err := r.redis.Keys(r.ctx, "access_count:*").Result()
	if err != nil {
		r.logger.WithError(err).Error("Failed to fetch access count keys from Redis")
		return
	}
	for _, key := range keys {
		shortCode := key[len("access_count:"):]
		count, err := r.redis.Get(r.ctx, key).Int()

		if err != nil {
			r.logger.WithError(err).WithField("short_code", shortCode).Warn("Failed to fetch access count from Redis")
			continue
		}
		if count > 0 {
			_, err := r.db.Exec(
				"UPDATE urls SET access_count = access_count + $1 WHERE short_code $2",
				count, shortCode,
			)
			if err == nil {
				r.logger.WithFields(logrus.Fields{
					"short_code": shortCode,
					"count":      count,
				}).Info("Synced access count to database")
				r.redis.Del(r.ctx, key)
			} else {
				r.logger.WithError(err).WithField("short_code", shortCode).Error("Failed to sync access count to database")
			}
		}
	}
}
