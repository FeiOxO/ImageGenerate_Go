package middleware

import (
	"net/http"
	"strings"

	"ai-image-demo-backend/internal/config"
	"ai-image-demo-backend/internal/models"
	"ai-image-demo-backend/internal/utils"
)

func Auth(cfg config.Config, next func(http.ResponseWriter, *http.Request, models.User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			utils.Error(w, http.StatusUnauthorized, "missing authorization token")
			return
		}

		claims, err := utils.ParseJWT(cfg.JWTSecret, strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			utils.Error(w, http.StatusUnauthorized, "invalid authorization token")
			return
		}

		user := models.User{
			ID:       claims.UserID,
			Username: claims.Username,
		}
		next(w, WithUser(r, user), user)
	}
}
