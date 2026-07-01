package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"ai-image-demo-backend/internal/config"
	"ai-image-demo-backend/internal/models"
	"ai-image-demo-backend/internal/repositories"
	"ai-image-demo-backend/internal/utils"
)

type AuthService struct {
	users *repositories.UserRepository
	cfg   config.Config
}

type AuthResult struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func NewAuthService(users *repositories.UserRepository, cfg config.Config) *AuthService {
	return &AuthService{users: users, cfg: cfg}
}

func (s *AuthService) Register(ctx context.Context, phone string, username string, password string) (models.User, error) {
	phone = strings.TrimSpace(phone)
	username = strings.TrimSpace(username)
	if phone == "" || username == "" || password == "" {
		return models.User{}, errors.New("phone, username and password are required")
	}
	if len(password) < 6 {
		return models.User{}, errors.New("password must contain at least 6 characters")
	}

	exists, err := s.users.ExistsPhone(ctx, phone)
	if err != nil {
		return models.User{}, err
	}
	if exists {
		return models.User{}, errors.New("phone already exists")
	}

	exists, err = s.users.ExistsUsername(ctx, username)
	if err != nil {
		return models.User{}, err
	}
	if exists {
		return models.User{}, errors.New("username already exists")
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return models.User{}, err
	}

	user := models.User{
		Phone:        phone,
		Username:     username,
		PasswordHash: hash,
	}
	if err := s.users.Create(ctx, &user); err != nil {
		return models.User{}, err
	}
	user.PasswordHash = ""
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, account string, password string) (AuthResult, error) {
	account = strings.TrimSpace(account)
	if account == "" || password == "" {
		return AuthResult{}, errors.New("account and password are required")
	}

	user, err := s.users.FindByAccount(ctx, account)
	if errors.Is(err, sql.ErrNoRows) {
		return AuthResult{}, errors.New("invalid account or password")
	}
	if err != nil {
		return AuthResult{}, err
	}
	if !utils.CheckPassword(user.PasswordHash, password) {
		return AuthResult{}, errors.New("invalid account or password")
	}

	token, err := utils.SignJWT(s.cfg.JWTSecret, s.cfg.JWTExpireHours, user.ID, user.Username)
	if err != nil {
		return AuthResult{}, err
	}

	user.PasswordHash = ""
	return AuthResult{Token: token, User: user}, nil
}

func (s *AuthService) FindMe(ctx context.Context, id int64) (models.User, error) {
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return models.User{}, err
	}
	user.PasswordHash = ""
	return user, nil
}
