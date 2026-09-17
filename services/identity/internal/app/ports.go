package app

import (
	"context"

	"golangFoodService/services/identity/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	GetByID(ctx context.Context, id string) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Count(ctx context.Context) (int, error)
}
