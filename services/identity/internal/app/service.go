package app

import (
	"context"
	"errors"
	"time"

	"golangFoodService/internal/platform/apperr"
	"golangFoodService/internal/platform/authx"
	"golangFoodService/internal/platform/id"
	"golangFoodService/services/identity/internal/domain"
)

type Service struct {
	users UserRepository
	jwt   *authx.JWT
}

func NewService(users UserRepository, jwt *authx.JWT) *Service {
	return &Service{users: users, jwt: jwt}
}

type AuthResult struct {
	User   domain.User
	Tokens authx.Tokens
}

func (s *Service) Register(ctx context.Context, email, password string, role authx.Role) (AuthResult, error) {
	email, err := domain.NormalizeEmail(email)
	if err != nil {
		return AuthResult{}, err
	}
	if err := domain.ValidatePassword(password); err != nil {
		return AuthResult{}, err
	}
	if role == "" {
		role = authx.RoleCustomer
	}
	if _, err := authx.ParseRole(string(role)); err != nil {
		return AuthResult{}, err
	}

	if _, err := s.users.GetByEmail(ctx, email); err == nil {
		return AuthResult{}, apperr.Conflict("email already registered")
	} else if !errors.Is(err, apperr.ErrNotFound) {
		var app *apperr.AppError
		if errors.As(err, &app) && errors.Is(app.Kind, apperr.ErrNotFound) {
			// continue
		} else {
			return AuthResult{}, err
		}
	}

	hash, err := authx.HashPassword(password)
	if err != nil {
		return AuthResult{}, apperr.Internal(err)
	}

	user := domain.User{
		ID:           id.New(),
		Email:        email,
		PasswordHash: hash,
		Role:         role,
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.users.Create(ctx, user); err != nil {
		return AuthResult{}, err
	}
	return s.issue(user)
}

func (s *Service) Login(ctx context.Context, email, password string) (AuthResult, error) {
	email, err := domain.NormalizeEmail(email)
	if err != nil {
		return AuthResult{}, err
	}
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return AuthResult{}, apperr.Unauthorized("invalid credentials")
	}
	if err := authx.VerifyPassword(password, user.PasswordHash); err != nil {
		return AuthResult{}, err
	}
	return s.issue(user)
}

func (s *Service) Refresh(_ context.Context, refreshToken string) (AuthResult, error) {
	claims, err := s.jwt.ParseRefresh(refreshToken)
	if err != nil {
		return AuthResult{}, err
	}
	user, err := s.users.GetByID(context.Background(), claims.Subject)
	if err != nil {
		return AuthResult{}, apperr.Unauthorized("invalid token")
	}
	return s.issue(user)
}

func (s *Service) GetUser(ctx context.Context, id string) (domain.User, error) {
	if id == "" {
		return domain.User{}, apperr.Invalid("id is required")
	}
	return s.users.GetByID(ctx, id)
}

func (s *Service) issue(user domain.User) (AuthResult, error) {
	tokens, err := s.jwt.Issue(user.ID, user.Email, user.Role)
	if err != nil {
		return AuthResult{}, apperr.Internal(err)
	}
	return AuthResult{User: user, Tokens: tokens}, nil
}
