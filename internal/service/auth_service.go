package service

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"inteam/internal/auth"
	"inteam/internal/domain"
	"inteam/internal/repository"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (*domain.AuthUser, error)
	Login(ctx context.Context, email, password string) (string, string, error)
	GetUserByID(ctx context.Context, id uint) (*domain.AuthUser, error)
}

type authService struct {
	users      repository.UserRepository
	jwtManager *auth.JWTManager
	logger     *zap.Logger
}

func NewAuthService(
	users repository.UserRepository,
	jwtManager *auth.JWTManager,
	logger *zap.Logger,
) AuthService {
	return &authService{
		users:      users,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

func (s *authService) Register(ctx context.Context, email, password string) (*domain.AuthUser, error) {
	existing, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("user already exists")
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &domain.AuthUser{
		Email:        email,
		PasswordHash: hash,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", errors.New("invalid credentials")
	}

	if err := auth.CheckPassword(password, user.PasswordHash); err != nil {
		return "", "", errors.New("invalid credentials")
	}

	access, refresh, err := s.jwtManager.GenerateTokens(user.ID)
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func (s *authService) GetUserByID(ctx context.Context, id uint) (*domain.AuthUser, error) {
	return s.users.GetByID(ctx, id)
}

