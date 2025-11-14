package repository

import (
	"context"

	"gorm.io/gorm"

	"inteam/internal/domain"
)

type ProfileRepository interface {
	GetByVKID(ctx context.Context, vkID int64) (*domain.Profile, error)
	Save(ctx context.Context, profile *domain.Profile) error
}

type profileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) ProfileRepository {
	return &profileRepository{db: db}
}

func (r *profileRepository) GetByVKID(ctx context.Context, vkID int64) (*domain.Profile, error) {
	var profile domain.Profile
	if err := r.db.WithContext(ctx).Where("vkid = ?", vkID).First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

func (r *profileRepository) Save(ctx context.Context, profile *domain.Profile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

