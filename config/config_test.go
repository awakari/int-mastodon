package config

import (
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	os.Setenv("API_HTTP_HOST", "activitypub.awakari.com")
	os.Setenv("API_WRITER_BACKOFF", "23h")
	os.Setenv("API_WRITER_URI", "writer:56789")
	os.Setenv("LOG_LEVEL", "4")
	os.Setenv("API_KEY_PUBLIC", "xxx")
	os.Setenv("API_KEY_PRIVATE", "yyy")
	os.Setenv("API_MASTODON_CLIENT_KEYS", "key1,key2")
	os.Setenv("API_MASTODON_CLIENT_SECRETS", "secret1,secret2")
	os.Setenv("API_MASTODON_CLIENT_TOKENS", "token1,token2")
	cfg, err := NewConfigFromEnv()
	assert.Nil(t, err)
	assert.Equal(t, 23*time.Hour, cfg.Api.Writer.Backoff)
	assert.Equal(t, "writer:56789", cfg.Api.Writer.Uri)
	assert.Equal(t, slog.LevelWarn, slog.Level(cfg.Log.Level))
	assert.Equal(t, []string{"key1", "key2"}, cfg.Api.Mastodon.Client.Keys)
}
