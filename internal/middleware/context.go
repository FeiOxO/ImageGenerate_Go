package middleware

import (
	"context"
	"net/http"

	"ai-image-demo-backend/internal/models"
)

type contextKey string

const userContextKey contextKey = "user"

func WithUser(r *http.Request, user models.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func CurrentUser(r *http.Request) (models.User, bool) {
	user, ok := r.Context().Value(userContextKey).(models.User)
	return user, ok
}
