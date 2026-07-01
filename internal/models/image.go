package models

import "time"

type GeneratedImage struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"userId"`
	Prompt       string    `json:"prompt"`
	ImagePath    string    `json:"imagePath"`
	Status       string    `json:"status"`
	DurationMS   int64     `json:"durationMs"`
	ErrorMessage string    `json:"errorMessage,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
