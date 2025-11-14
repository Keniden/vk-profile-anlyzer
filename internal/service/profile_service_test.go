package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"inteam/internal/domain"
)

type vkClientMock struct {
	user    *domain.VKUser
	wall    []domain.WallPost
	gifts   []domain.Gift
	friends []domain.Friend
	err     error
}

func (m *vkClientMock) GetUser(ctx context.Context, vkID int64) (*domain.VKUser, error) {
	return m.user, m.err
}

func (m *vkClientMock) GetWall(ctx context.Context, vkID int64, offset, count int) ([]domain.WallPost, error) {
	return m.wall, m.err
}

func (m *vkClientMock) GetGifts(ctx context.Context, vkID int64, offset, count int) ([]domain.Gift, error) {
	return m.gifts, m.err
}

func (m *vkClientMock) GetFriends(ctx context.Context, vkID int64, offset, count int) ([]domain.Friend, error) {
	return m.friends, m.err
}

type gigachatMock struct {
	summary string
	err     error
}

func (g *gigachatMock) GenerateProfileSummary(ctx context.Context, data domain.ProfileData) (string, error) {
	return g.summary, g.err
}

type profileRepoMock struct {
	saved *domain.Profile
	err   error
}

func (r *profileRepoMock) GetByVKID(ctx context.Context, vkID int64) (*domain.Profile, error) {
	return nil, nil
}

func (r *profileRepoMock) Save(ctx context.Context, profile *domain.Profile) error {
	r.saved = profile
	return r.err
}

func TestAnalyzeProfile_Basic(t *testing.T) {
	vkMock := &vkClientMock{
		user: &domain.VKUser{
			ID:        1,
			FirstName: "Test",
			LastName:  "User",
		},
		wall: []domain.WallPost{
			{Text: "hello", Date: time.Unix(0, 0)},
			{Text: "world", Date: time.Unix(86400, 0)},
		},
	}
	ggMock := &gigachatMock{summary: "test summary"}
	repoMock := &profileRepoMock{}

	svc := &profileService{
		vkClient:    vkMock,
		gigachat:    ggMock,
		profileRepo: repoMock,
	}

	profile, err := svc.AnalyzeProfile(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, profile)
	require.Equal(t, "test summary", profile.Summary)
	require.NotEmpty(t, repoMock.saved.RawJSON)
}

