package vk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"

	"inteam/internal/config"
	"inteam/internal/domain"
)

type Client interface {
	GetUser(ctx context.Context, vkID int64) (*domain.VKUser, error)
	GetWall(ctx context.Context, vkID int64, offset, count int) ([]domain.WallPost, error)
	GetGifts(ctx context.Context, vkID int64, offset, count int) ([]domain.Gift, error)
	GetFriends(ctx context.Context, vkID int64, offset, count int) ([]domain.Friend, error)
}

type client struct {
	cfg        config.VKConfig
	httpClient *http.Client
	logger     *zap.Logger
	cache      sync.Map
	redis      redis.UniversalClient
	cb         *gobreaker.CircuitBreaker
}

func NewClient(cfg config.VKConfig, httpClient *http.Client, logger *zap.Logger, redisClient redis.UniversalClient) Client {
	return &client{
		cfg:        cfg,
		httpClient: httpClient,
		logger:     logger,
		redis:      redisClient,
		cb: gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        "vk_api",
			MaxRequests: 5,
			Interval:    30 * time.Second,
			Timeout:     30 * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures >= 5
			},
		}),
	}
}

type vkResponse struct {
	Response json.RawMessage `json:"response"`
	Error    *struct {
		ErrorCode int    `json:"error_code"`
		ErrorMsg  string `json:"error_msg"`
	} `json:"error,omitempty"`
}

func (c *client) GetUser(ctx context.Context, vkID int64) (*domain.VKUser, error) {
	cacheKey := fmt.Sprintf("user:%d", vkID)

	if c.redis != nil {
		if val, err := c.redis.Get(ctx, cacheKey).Result(); err == nil {
			var user domain.VKUser
			if err := json.Unmarshal([]byte(val), &user); err == nil {
				return &user, nil
			}
		}
	}

	if val, ok := c.cache.Load(cacheKey); ok {
		if user, ok := val.(*domain.VKUser); ok {
			return user, nil
		}
	}

	params := url.Values{}
	params.Set("user_ids", strconv.FormatInt(vkID, 10))
	params.Set("fields", "bdate,city,about,sex,screen_name")

	var users []struct {
		ID         int64  `json:"id"`
		ScreenName string `json:"screen_name"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		Sex        int    `json:"sex"`
		BDate      string `json:"bdate"`
		City       struct {
			Title string `json:"title"`
		} `json:"city"`
		About string `json:"about"`
	}

	if err := c.callVK(ctx, "users.get", params, &users); err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	u := users[0]
	user := &domain.VKUser{
		ID:         u.ID,
		ScreenName: u.ScreenName,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		Sex:        u.Sex,
		BirthDate:  u.BDate,
		City:       u.City.Title,
		About:      u.About,
	}

	c.cache.Store(cacheKey, user)

	if c.redis != nil {
		if b, err := json.Marshal(user); err == nil {
			_ = c.redis.Set(ctx, cacheKey, b, 10*time.Minute).Err()
		}
	}
	return user, nil
}

func (c *client) GetWall(ctx context.Context, vkID int64, offset, count int) ([]domain.WallPost, error) {
	params := url.Values{}
	params.Set("owner_id", strconv.FormatInt(vkID, 10))
	params.Set("offset", strconv.Itoa(offset))
	params.Set("count", strconv.Itoa(count))

	var resp struct {
		Count int `json:"count"`
		Items []struct {
			ID       int64  `json:"id"`
			Date     int64  `json:"date"`
			Text     string `json:"text"`
			Likes    struct {
				Count int `json:"count"`
			} `json:"likes"`
			Reposts struct {
				Count int `json:"count"`
			} `json:"reposts"`
			Comments struct {
				Count int `json:"count"`
			} `json:"comments"`
			Views struct {
				Count int `json:"count"`
			} `json:"views"`
			PostType string `json:"post_type"`
			IsPinned int    `json:"is_pinned"`
		} `json:"items"`
	}

	if err := c.callVK(ctx, "wall.get", params, &resp); err != nil {
		return nil, err
	}

	posts := make([]domain.WallPost, 0, len(resp.Items))
	for _, p := range resp.Items {
		posts = append(posts, domain.WallPost{
			ID:       p.ID,
			Date:     time.Unix(p.Date, 0),
			Text:     p.Text,
			Likes:    p.Likes.Count,
			Reposts:  p.Reposts.Count,
			Comments: p.Comments.Count,
			Views:    p.Views.Count,
			PostType: p.PostType,
			IsPinned: p.IsPinned == 1,
		})
	}
	return posts, nil
}

func (c *client) GetGifts(ctx context.Context, vkID int64, offset, count int) ([]domain.Gift, error) {
	params := url.Values{}
	params.Set("user_id", strconv.FormatInt(vkID, 10))
	params.Set("offset", strconv.Itoa(offset))
	params.Set("count", strconv.Itoa(count))

	var resp struct {
		Count int `json:"count"`
		Items []struct {
			ID   int64  `json:"id"`
			Text string `json:"text"`
		} `json:"items"`
	}

	if err := c.callVK(ctx, "gifts.get", params, &resp); err != nil {
		return nil, err
	}

	gifts := make([]domain.Gift, 0, len(resp.Items))
	for _, g := range resp.Items {
		gifts = append(gifts, domain.Gift{
			ID:   g.ID,
			Text: g.Text,
		})
	}
	return gifts, nil
}

func (c *client) GetFriends(ctx context.Context, vkID int64, offset, count int) ([]domain.Friend, error) {
	params := url.Values{}
	params.Set("user_id", strconv.FormatInt(vkID, 10))
	params.Set("offset", strconv.Itoa(offset))
	params.Set("count", strconv.Itoa(count))
	params.Set("fields", "sex")

	var resp struct {
		Count int `json:"count"`
		Items []struct {
			ID        int64  `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Sex       int    `json:"sex"`
		} `json:"items"`
	}

	if err := c.callVK(ctx, "friends.get", params, &resp); err != nil {
		return nil, err
	}

	friends := make([]domain.Friend, 0, len(resp.Items))
	for _, f := range resp.Items {
		friends = append(friends, domain.Friend{
			ID:        f.ID,
			FirstName: f.FirstName,
			LastName:  f.LastName,
			Sex:       f.Sex,
		})
	}
	return friends, nil
}

func (c *client) callVK(ctx context.Context, method string, params url.Values, out interface{}) error {
	endpoint := fmt.Sprintf("%s/%s", c.cfg.BaseURL, method)

	params.Set("access_token", c.cfg.AccessToken)
	params.Set("v", c.cfg.APIVersion)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.URL.RawQuery = params.Encode()

	operation := func() (interface{}, error) {
		var lastErr error
		for attempt := 0; attempt < 3; attempt++ {
			resp, err := c.httpClient.Do(req)
			if err != nil {
				lastErr = err
				c.logger.Warn("vk request error", zap.Error(err), zap.Int("attempt", attempt+1))
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}

			defer resp.Body.Close()

			if resp.StatusCode == http.StatusTooManyRequests {
				c.logger.Warn("vk rate limited", zap.Int("attempt", attempt+1))
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}

			if resp.StatusCode >= 400 {
				return nil, fmt.Errorf("vk api error: status=%d", resp.StatusCode)
			}

			var vkResp vkResponse
			if err := json.NewDecoder(resp.Body).Decode(&vkResp); err != nil {
				return nil, err
			}

			if vkResp.Error != nil {
				return nil, fmt.Errorf("vk error: code=%d msg=%s", vkResp.Error.ErrorCode, vkResp.Error.ErrorMsg)
			}

			if out != nil {
				if err := json.Unmarshal(vkResp.Response, out); err != nil {
					return nil, err
				}
			}

			return nil, nil
		}

		return nil, lastErr
	}

	_, err = c.cb.Execute(operation)
	return err
}
