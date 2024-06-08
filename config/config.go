package config

import (
	"github.com/kelseyhightower/envconfig"
	"time"
)

type Config struct {
	Api struct {
		Port   uint16 `envconfig:"API_PORT" default:"50051" required:"true"`
		Writer struct {
			Backoff   time.Duration `envconfig:"API_WRITER_BACKOFF" default:"10s" required:"true"`
			BatchSize uint32        `envconfig:"API_WRITER_BATCH_SIZE" default:"16" required:"true"`
			Uri       string        `envconfig:"API_WRITER_URI" default:"resolver:50051" required:"true"`
		}
		ActivityPub struct {
			Uri string `envconfig:"API_ACTIVITYPUB_URI" default:"int-activitypub:50051" required:"true"`
		}
		Mastodon MastodonConfig
	}
	Log struct {
		Level int `envconfig:"LOG_LEVEL" default:"-4" required:"true"`
	}
}

type MastodonConfig struct {
	Client struct {
		Key       string `envconfig:"API_MASTODON_CLIENT_KEY" required:"true"`
		Secret    string `envconfig:"API_MASTODON_CLIENT_SECRET" required:"true"`
		Token     string `envconfig:"API_MASTODON_CLIENT_TOKEN" required:"true"`
		UserAgent string `envconfig:"API_MASTODON_USER_AGENT" default:"awakari" required:"true""`
	}
	Endpoint struct {
		Search string `envconfig:"API_MASTODON_ENDPOINT_SEARCH" default:"https://mastodon.social/api/v2/search" required:"true"`
		Stream string `envconfig:"API_MASTODON_ENDPOINT_STREAM" default:"https://streaming.mastodon.social/api/v1/streaming/public?remote=false&only_media=false" required:"true"`
	}
	StreamTimeoutMax time.Duration `envconfig:"API_MASTODON_STREAM_TIMEOUT_MAX" default:"1m" required:"true"`
}

func NewConfigFromEnv() (cfg Config, err error) {
	err = envconfig.Process("", &cfg)
	return
}
