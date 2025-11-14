package gigachat

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"inteam/internal/config"
	"inteam/internal/domain"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestGenerateProfileSummary_Basic(t *testing.T) {
	client := &http.Client{
		Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
			body := `{"text":"summary"}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	logger, _ := zap.NewDevelopment()
	cfg := config.GigaChatConfig{
		BaseURL: "https://gigachat.example.com",
		Token:   "test",
	}

	c := NewClient(cfg, client, logger)
	text, err := c.GenerateProfileSummary(context.Background(), domain.ProfileData{})
	require.NoError(t, err)
	require.Equal(t, "summary", text)
}

