package vk

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"inteam/internal/config"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func newTestHTTPClient(fn roundTripperFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func TestGetUser_Basic(t *testing.T) {
	client := newTestHTTPClient(func(r *http.Request) (*http.Response, error) {
		body := `{"response":[{"id":1,"screen_name":"test","first_name":"Test","last_name":"User","sex":2,"bdate":"1.1.2000","city":{"title":"City"},"about":"About"}]}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})

	logger, _ := zap.NewDevelopment()
	cfg := config.VKConfig{
		BaseURL:    "https://api.vk.com/method",
		APIVersion: "5.199",
	}

	c := NewClient(cfg, client, logger, nil)
	user, err := c.GetUser(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, int64(1), user.ID)
	require.Equal(t, "Test", user.FirstName)
}
