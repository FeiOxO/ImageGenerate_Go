package repositories

import (
	"context"
	"database/sql"
	"time"

	"ai-image-demo-backend/internal/models"
)

type ImageRepository struct {
	db *sql.DB
}

func NewImageRepository(db *sql.DB) *ImageRepository {
	return &ImageRepository{db: db}
}

func (r *ImageRepository) Create(ctx context.Context, image *models.GeneratedImage) error {
	now := time.Now()
	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO generated_images
		 (user_id, prompt, image_path, status, duration_ms, error_message, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		image.UserID,
		image.Prompt,
		image.ImagePath,
		image.Status,
		image.DurationMS,
		nullableString(image.ErrorMessage),
		now,
		now,
	)
	if err != nil {
		return err
	}

	image.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}
	image.CreatedAt = now
	image.UpdatedAt = now
	return nil
}

func (r *ImageRepository) ListByUser(ctx context.Context, userID int64, page int, pageSize int) ([]models.GeneratedImage, int64, error) {
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM generated_images WHERE user_id = ?`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, user_id, prompt, image_path, status, duration_ms, COALESCE(error_message, ''), created_at, updated_at
		 FROM generated_images
		 WHERE user_id = ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		userID,
		pageSize,
		offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	images := make([]models.GeneratedImage, 0)
	for rows.Next() {
		var image models.GeneratedImage
		if err := rows.Scan(
			&image.ID,
			&image.UserID,
			&image.Prompt,
			&image.ImagePath,
			&image.Status,
			&image.DurationMS,
			&image.ErrorMessage,
			&image.CreatedAt,
			&image.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		images = append(images, image)
	}

	return images, total, rows.Err()
}

func nullableString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}
