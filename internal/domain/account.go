package domain

import "time"

type Account struct {
	ID           int       `json:"user_id"`
	Name         string    `json:"user_name"`
	Email        string    `json:"email"`
	IsSeller     bool      `json:"is_seller"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	PasswordHash string    `json:"-"`
}
