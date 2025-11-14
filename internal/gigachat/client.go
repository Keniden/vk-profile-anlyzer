package gigachat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"inteam/internal/config"
	"inteam/internal/domain"
)

type Client interface {
	GenerateProfileSummary(ctx context.Context, data domain.ProfileData) (string, error)
}

type client struct {
	cfg        config.GigaChatConfig
	httpClient *http.Client
	logger     *zap.Logger
	cb         *gobreaker.CircuitBreaker
}

func NewClient(cfg config.GigaChatConfig, httpClient *http.Client, logger *zap.Logger) Client {
	return &client{
		cfg:        cfg,
		httpClient: httpClient,
		logger:     logger,
		cb: gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        "gigachat",
			MaxRequests: 5,
			Interval:    30 * time.Second,
			Timeout:     30 * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures >= 5
			},
		}),
	}
}

type requestBody struct {
	Prompt string `json:"prompt"`
}

type responseBody struct {
	Text string `json:"text"`
}

func (c *client) GenerateProfileSummary(ctx context.Context, data domain.ProfileData) (string, error) {
	tracer := otel.Tracer("inteam/client/gigachat")
	ctx, span := tracer.Start(ctx, "GenerateProfileSummary")
	span.SetAttributes(
		attribute.Int("wall_posts", len(data.Wall)),
		attribute.Int("friends", len(data.Friends)),
		attribute.Int("gifts", len(data.Gifts)),
	)
	defer span.End()

	prompt := buildPrompt(data)

	body, err := json.Marshal(requestBody{Prompt: prompt})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.cfg.Token))

	start := time.Now()

	operation := func() (interface{}, error) {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("gigachat error: status=%d", resp.StatusCode)
		}

		var respBody responseBody
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			return nil, err
		}
		return respBody.Text, nil
	}

	result, err := c.cb.Execute(operation)
	if err != nil {
		return "", err
	}

	text, _ := result.(string)
	if len(text) > 2000 {
		text = text[:2000]
	}

	c.logger.Info("gigachat call",
		zap.Duration("latency", time.Since(start)),
		zap.Int("summary_len", len(text)),
	)

	return text, nil
}

func buildPrompt(data domain.ProfileData) string {
	return fmt.Sprintf(
		`Ты — аналитик социальных сетей. 
Проанализируй профиль VK пользователя и кратко опиши основные черты личности, интересы и социальную активность в 5–7 предложениях на русском языке.

Основная информация:
- Имя: %s %s
- Город: %s
- О себе: %s
- Количество друзей: %d
- Количество подарков: %d

Активность на стене:
- Количество постов: %d
- Средняя длина поста: %.1f символов
- Средний уровень вовлеченности: %.2f
- Плотность активности (постов в месяц): %.2f

Сформируй человеческое, понятное резюме без упоминания технических деталей и метрик.`,
		data.User.FirstName,
		data.User.LastName,
		data.User.City,
		data.User.About,
		len(data.Friends),
		len(data.Gifts),
		len(data.Wall),
		data.Vector.AveragePostLen,
		data.Vector.EngagementRate,
		data.Vector.PostsPerMonth,
	)
}
