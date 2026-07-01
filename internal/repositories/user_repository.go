package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ai-image-demo-backend/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	now := time.Now()
	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO users (phone, username, password_hash, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)`,
		user.Phone,
		user.Username,
		user.PasswordHash,
		now,
		now,
	)
	if err != nil {
		return err
	}

	user.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}
	user.CreatedAt = now
	user.UpdatedAt = now
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (models.User, error) {
	return r.scanOne(ctx, `SELECT id, phone, username, password_hash, created_at, updated_at FROM users WHERE id = ?`, id)
}

func (r *UserRepository) FindByAccount(ctx context.Context, account string) (models.User, error) {
	return r.scanOne(ctx, `SELECT id, phone, username, password_hash, created_at, updated_at FROM users WHERE username = ? OR phone = ?`, account, account)
}

func (r *UserRepository) ExistsPhone(ctx context.Context, phone string) (bool, error) {
	return r.exists(ctx, `SELECT 1 FROM users WHERE phone = ? LIMIT 1`, phone)
}

func (r *UserRepository) ExistsUsername(ctx context.Context, username string) (bool, error) {
	return r.exists(ctx, `SELECT 1 FROM users WHERE username = ? LIMIT 1`, username)
}

func (r *UserRepository) scanOne(ctx context.Context, query string, args ...any) (models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.Phone,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *UserRepository) exists(ctx context.Context, query string, args ...any) (bool, error) {
	var marker int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&marker)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
