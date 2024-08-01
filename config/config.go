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
		Event struct {
			Type string `envconfig:"API_EVENT_TYPE" required:"true" default:"com.awakari.mastodon.v1"`
		}
		ActivityPub struct {
			Host string `envconfig:"API_ACTIVITYPUB_HOST" default:"activitypub.awakari.com" required:"true"`
			Uri  string `envconfig:"API_ACTIVITYPUB_URI" default:"int-activitypub:50051" required:"true"`
		}
		Mastodon MastodonConfig
		Queue    QueueConfig
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
		Accounts string `envconfig:"API_MASTODON_ENDPOINT_SEARCH" default:"https://mastodon.social/api/v1/accounts" required:"true"`
		Search   string `envconfig:"API_MASTODON_ENDPOINT_SEARCH" default:"https://mastodon.social/api/v2/search" required:"true"`
		Stream   string `envconfig:"API_MASTODON_ENDPOINT_STREAM" default:"https://streaming.mastodon.social/api/v1/streaming/public?remote=false&only_media=false" required:"true"`
	}
	StreamTimeoutMax time.Duration `envconfig:"API_MASTODON_STREAM_TIMEOUT_MAX" default:"5m" required:"true"`
	CountMin         struct {
		Followers uint32 `envconfig:"API_MASTODON_COUNT_MIN_FOLLOWERS" default:"100" required:"true"`
		Posts     uint32 `envconfig:"API_MASTODON_COUNT_MIN_POSTS" default:"1000" required:"true"`
	}
	Search struct {
		Limit uint32 `envconfig:"API_MASTODON_SEARCH_LIMIT" default:"5" required:"true"`
	}
}

type QueueConfig struct {
	Uri              string `envconfig:"API_QUEUE_URI" default:"queue:50051" required:"true"`
	InterestsCreated struct {
		BatchSize uint32 `envconfig:"API_QUEUE_INTERESTS_CREATED_BATCH_SIZE" default:"1" required:"true"`
		Name      string `envconfig:"API_QUEUE_INTERESTS_CREATED_NAME" default:"source-search" required:"true"`
		Subj      string `envconfig:"API_QUEUE_INTERESTS_CREATED_SUBJ" default:"interests-created" required:"true"`
	}
	InterestsUpdated struct {
		BatchSize uint32 `envconfig:"API_QUEUE_INTERESTS_UPDATED_BATCH_SIZE" default:"1" required:"true"`
		Name      string `envconfig:"API_QUEUE_INTERESTS_UPDATED_NAME" default:"source-search" required:"true"`
		Subj      string `envconfig:"API_QUEUE_INTERESTS_UPDATED_SUBJ" default:"interests-updated" required:"true"`
	}
}

func NewConfigFromEnv() (cfg Config, err error) {
	err = envconfig.Process("", &cfg)
	return
}
