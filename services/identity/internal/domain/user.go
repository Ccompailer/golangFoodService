package domain

import (
	"strings"
	"time"

	"golangFoodService/internal/platform/apperr"
	"golangFoodService/internal/platform/authx"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         authx.Role
	CreatedAt    time.Time
}

func NormalizeEmail(email string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || !strings.Contains(email, "@") || len(email) > 254 {
		return "", apperr.Invalid("invalid email")
	}
	return email, nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return apperr.Invalid("password must be at least 8 characters")
	}
	if len(password) > 72 {
		return apperr.Invalid("password is too long")
	}
	return nil
}
