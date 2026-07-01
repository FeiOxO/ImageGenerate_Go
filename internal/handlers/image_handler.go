package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ai-image-demo-backend/internal/models"
	"ai-image-demo-backend/internal/repositories"
	"ai-image-demo-backend/internal/services"
	"ai-image-demo-backend/internal/utils"
)

const maxGenerateCount = 50

type ImageHandler struct {
	images *repositories.ImageRepository
	pool   *services.WorkerPool
}

func NewImageHandler(images *repositories.ImageRepository, pool *services.WorkerPool) *ImageHandler {
	return &ImageHandler{images: images, pool: pool}
}

func (h *ImageHandler) Generate(w http.ResponseWriter, r *http.Request, user models.User) {
	var req generateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Prompt = strings.TrimSpace(req.Prompt)
	if req.Prompt == "" {
		utils.Error(w, http.StatusBadRequest, "prompt is required")
		return
	}
	if req.Count < 1 {
		req.Count = 1
	}
	if req.Count > maxGenerateCount {
		utils.Error(w, http.StatusBadRequest, "count cannot exceed 50")
		return
	}

	results := make([]chan services.ImageResult, 0, req.Count)
	for i := 0; i < req.Count; i++ {
		result := make(chan services.ImageResult, 1)
		if err := h.pool.Submit(services.ImageJob{
			UserID: user.ID,
			Prompt: req.Prompt,
			Result: result,
		}); err != nil {
			utils.Error(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		results = append(results, result)
	}

	items := make([]models.GeneratedImage, 0, req.Count)
	timeout := time.After(10 * time.Minute)
	for _, result := range results {
		select {
		case value := <-result:
			items = append(items, value.Image)
		case <-timeout:
			utils.Error(w, http.StatusGatewayTimeout, "image generation timeout")
			return
		case <-r.Context().Done():
			utils.Error(w, http.StatusRequestTimeout, "request cancelled")
			return
		}
	}

	utils.JSON(w, http.StatusOK, map[string][]models.GeneratedImage{"items": items})
}

func (h *ImageHandler) List(w http.ResponseWriter, r *http.Request, user models.User) {
	page := parsePositiveInt(r.URL.Query().Get("page"), 1)
	pageSize := parsePositiveInt(r.URL.Query().Get("pageSize"), 20)
	if pageSize > 100 {
		pageSize = 100
	}

	items, total, err := h.images.ListByUser(r.Context(), user.ID, page, pageSize)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]any{
		"items":    items,
		"page":     page,
		"pageSize": pageSize,
		"total":    total,
	})
}

func parsePositiveInt(value string, fallback int) int {
	number, err := strconv.Atoi(value)
	if err != nil || number < 1 {
		return fallback
	}
	return number
}

type generateRequest struct {
	Prompt string `json:"prompt"`
	Count  int    `json:"count"`
}
