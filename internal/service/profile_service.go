package service

import (
	"context"
	"encoding/json"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"inteam/internal/domain"
	"inteam/internal/gigachat"
	"inteam/internal/repository"
	"inteam/internal/storage"
	"inteam/internal/vk"
)

type ProfileService interface {
	GetProfile(ctx context.Context, vkID int64) (*domain.Profile, error)
	AnalyzeProfile(ctx context.Context, vkID int64) (*domain.Profile, error)
}

type profileService struct {
	vkClient    vk.Client
	gigachat    gigachat.Client
	profileRepo repository.ProfileRepository
	storage     storage.ObjectStorage
	logger      *zap.Logger
}

func NewProfileService(
	vkClient vk.Client,
	gigachat gigachat.Client,
	profileRepo repository.ProfileRepository,
	storage storage.ObjectStorage,
	logger *zap.Logger,
) ProfileService {
	return &profileService{
		vkClient:    vkClient,
		gigachat:    gigachat,
		profileRepo: profileRepo,
		storage:     storage,
		logger:      logger,
	}
}

func (s *profileService) GetProfile(ctx context.Context, vkID int64) (*domain.Profile, error) {
	return s.profileRepo.GetByVKID(ctx, vkID)
}

func (s *profileService) AnalyzeProfile(ctx context.Context, vkID int64) (*domain.Profile, error) {
	tracer := otel.Tracer("inteam/service/profile")
	ctx, span := tracer.Start(ctx, "AnalyzeProfile")
	span.SetAttributes(attribute.Int64("vk.id", vkID))
	defer span.End()

	user, err := s.vkClient.GetUser(ctx, vkID)
	if err != nil {
		return nil, err
	}

	wall, err := s.vkClient.GetWall(ctx, vkID, 0, 100)
	if err != nil {
		return nil, err
	}

	gifts, err := s.vkClient.GetGifts(ctx, vkID, 0, 100)
	if err != nil {
		return nil, err
	}

	friends, err := s.vkClient.GetFriends(ctx, vkID, 0, 100)
	if err != nil {
		return nil, err
	}

	vector := buildActivityVector(wall, gifts, friends)

	data := domain.ProfileData{
		User:    *user,
		Wall:    wall,
		Gifts:   gifts,
		Friends: friends,
		Vector:  vector,
	}

	summary, err := s.gigachat.GenerateProfileSummary(ctx, data)
	if err != nil {
		return nil, err
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	fullName := user.FirstName + " " + user.LastName

	profile := &domain.Profile{
		VKID:       vkID,
		ScreenName: user.ScreenName,
		FullName:   fullName,
		RawJSON:    string(raw),
		Summary:    summary,
		UpdatedAt:  time.Now(),
	}

	if err := s.profileRepo.Save(ctx, profile); err != nil {
		return nil, err
	}

	if s.storage != nil {
		if err := s.storage.SaveProfileSnapshot(ctx, vkID, raw); err != nil {
			s.logger.Warn("failed to save profile snapshot to object storage", zap.Error(err))
		}
	}

	return profile, nil
}

func buildActivityVector(wall []domain.WallPost, gifts []domain.Gift, friends []domain.Friend) domain.ActivityVector {
	if len(wall) == 0 {
		return domain.ActivityVector{
			PostsPerMonth:       0,
			AveragePostLen:      0,
			EngagementRate:      0,
			GiftsCount:          len(gifts),
			FriendsCount:        len(friends),
			ProfileCompleteness: 0,
		}
	}

	var (
		minDate  = wall[0].Date
		maxDate  = wall[0].Date
		totalLen int
		totalEng int
	)

	for _, p := range wall {
		if p.Date.Before(minDate) {
			minDate = p.Date
		}
		if p.Date.After(maxDate) {
			maxDate = p.Date
		}
		totalLen += len(p.Text)
		totalEng += p.Likes + p.Comments + p.Reposts
	}

	months := maxDate.Sub(minDate).Hours() / (24 * 30)
	if months < 1 {
		months = 1
	}

	postsPerMonth := float64(len(wall)) / months
	avgLen := float64(totalLen) / float64(len(wall))
	engagement := float64(totalEng) / float64(len(wall))

	return domain.ActivityVector{
		PostsPerMonth:       postsPerMonth,
		AveragePostLen:      avgLen,
		EngagementRate:      engagement,
		GiftsCount:          len(gifts),
		FriendsCount:        len(friends),
		ProfileCompleteness: 0,
	}
}
