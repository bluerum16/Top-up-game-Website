package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `db:"id"            json:"id"`
	Email        string     `db:"email"         json:"email"`
	Phone        string     `db:"phone"         json:"phone,omitempty"`
	Username     string     `db:"username"      json:"username"`
	PasswordHash string     `db:"password_hash" json:"-"`
	FullName     string     `db:"full_name"     json:"full_name,omitempty"`
	AvatarURL    string     `db:"avatar_url"    json:"avatar_url,omitempty"`
	Role         string     `db:"role"          json:"role"`
	Status       string     `db:"status"        json:"status"`
	LastLoginAt  *time.Time `db:"last_login_at" json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `db:"created_at"    json:"created_at"`
}

type RegisterRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
