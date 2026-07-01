package handlers

import (
	"net/http"

	"ai-image-demo-backend/internal/utils"
)

func Health(w http.ResponseWriter, _ *http.Request) {
	utils.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
