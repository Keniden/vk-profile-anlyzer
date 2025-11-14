package repository

import (
	"context"

	"gorm.io/gorm"

	"inteam/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.AuthUser) error
	GetByEmail(ctx context.Context, email string) (*domain.AuthUser, error)
	GetByID(ctx context.Context, id uint) (*domain.AuthUser, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.AuthUser) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.AuthUser, error) {
	var user domain.AuthUser
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*domain.AuthUser, error) {
	var user domain.AuthUser
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

