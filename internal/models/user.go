package models

import "time"

type User struct {
	ID           int64     `json:"id"`
	Phone        string    `json:"phone"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
