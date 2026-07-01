package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ai-image-demo-backend/internal/config"
	"ai-image-demo-backend/internal/models"
	"ai-image-demo-backend/internal/repositories"
)

type ImageService struct {
	images *repositories.ImageRepository
	cfg    config.Config
	logger *slog.Logger
	client *http.Client
}

type ImageJob struct {
	UserID int64
	Prompt string
	Result chan ImageResult
}

type ImageResult struct {
	Image models.GeneratedImage
	Err   error
}

func NewImageService(images *repositories.ImageRepository, cfg config.Config, logger *slog.Logger) *ImageService {
	return &ImageService{
		images: images,
		cfg:    cfg,
		logger: logger,
		client: &http.Client{Timeout: 10 * time.Minute},
	}
}

func (s *ImageService) Generate(ctx context.Context, job ImageJob) ImageResult {
	start := time.Now()
	image := models.GeneratedImage{
		UserID: job.UserID,
		Prompt: job.Prompt,
		Status: "failed",
	}

	s.logger.Info("image generation started", "user_id", job.UserID, "prompt_len", len(job.Prompt))

	imageBytes, err := s.callImageAPI(ctx, job.Prompt)
	image.DurationMS = time.Since(start).Milliseconds()
	if err != nil {
		image.ErrorMessage = err.Error()
		_ = s.images.Create(context.Background(), &image)
		s.logger.Error("image generation failed", "user_id", job.UserID, "duration_ms", image.DurationMS, "error", err)
		return ImageResult{Image: image, Err: err}
	}

	imagePath, err := s.savePNG(job.UserID, imageBytes)
	image.DurationMS = time.Since(start).Milliseconds()
	if err != nil {
		image.ErrorMessage = err.Error()
		_ = s.images.Create(context.Background(), &image)
		s.logger.Error("image save failed", "user_id", job.UserID, "duration_ms", image.DurationMS, "error", err)
		return ImageResult{Image: image, Err: err}
	}

	image.ImagePath = imagePath
	image.Status = "success"
	if err := s.images.Create(context.Background(), &image); err != nil {
		image.ErrorMessage = err.Error()
		image.Status = "failed"
		s.logger.Error("image database insert failed", "user_id", job.UserID, "error", err)
		return ImageResult{Image: image, Err: err}
	}

	s.logger.Info("image generation completed", "user_id", job.UserID, "duration_ms", image.DurationMS, "image_path", image.ImagePath)
	return ImageResult{Image: image}
}

func (s *ImageService) callImageAPI(ctx context.Context, prompt string) ([]byte, error) {
	body := map[string]any{
		"model":         "gpt-image-2",
		"prompt":        prompt,
		"n":             1,
		"quality":       "auto",
		"output_format": "png",
		"size":          "1024x1024",
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.ImageAPIBaseURL+"/v1/images/generations", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.ImageAPIKey)
	if s.cfg.ImageAPIAppID != "" {
		req.Header.Set("X-App-ID", s.cfg.ImageAPIAppID)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("image api http %d: %s", resp.StatusCode, truncate(string(responseBody), 1000))
	}

	var parsed imageAPIResponse
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return nil, fmt.Errorf("image api response is not json: %w", err)
	}

	imageBytes, err := s.extractImageBytes(ctx, parsed)
	if err != nil {
		return nil, err
	}
	if len(imageBytes) == 0 {
		return nil, errors.New("image api response has no image data")
	}
	return imageBytes, nil
}

func (s *ImageService) extractImageBytes(ctx context.Context, payload imageAPIResponse) ([]byte, error) {
	for _, item := range payload.Data {
		for _, value := range []string{item.B64JSON, item.Base64, item.ImageB64} {
			if value == "" {
				continue
			}
			return base64.StdEncoding.DecodeString(value)
		}

		if item.URL != "" {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, item.URL, nil)
			if err != nil {
				return nil, err
			}
			req.Header.Set("Authorization", "Bearer "+s.cfg.ImageAPIKey)
			resp, err := s.client.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return nil, fmt.Errorf("download generated image failed: http %d", resp.StatusCode)
			}
			return io.ReadAll(resp.Body)
		}
	}

	return nil, errors.New("image api response has no image data")
}

func (s *ImageService) savePNG(userID int64, imageBytes []byte) (string, error) {
	userDirName := fmt.Sprintf("user_%d", userID)
	userDir := filepath.Join(s.cfg.ImageOutputDir, userDirName)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s.png", time.Now().Format("20060102_150405_000000000"))
	fullPath := filepath.Join(userDir, filename)
	if err := os.WriteFile(fullPath, imageBytes, 0644); err != nil {
		return "", err
	}

	return "/storage/images/" + userDirName + "/" + filename, nil
}

type imageAPIResponse struct {
	Data []struct {
		B64JSON  string `json:"b64_json"`
		Base64   string `json:"base64"`
		ImageB64 string `json:"image_b64"`
		URL      string `json:"url"`
	} `json:"data"`
}

func truncate(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}
