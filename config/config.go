package config

import (
	"github.com/kelseyhightower/envconfig"
	"time"
)

type Config struct {
	Api struct {
		Port   uint16 `envconfig:"API_PORT" default:"50051" required:"true"`
		Writer struct {
			Backoff time.Duration `envconfig:"API_WRITER_BACKOFF" default:"10s" required:"true"`
			Timeout time.Duration `envconfig:"API_WRITER_TIMEOUT" default:"10s" required:"true"`
			Uri     string        `envconfig:"API_WRITER_URI" default:"http://pub:8080/v1" required:"true"`
		}
		Event struct {
			Type string `envconfig:"API_EVENT_TYPE" required:"true" default:"com_awakari_mastodon_v1"`
		}
		ActivityPub struct {
			Host string `envconfig:"API_ACTIVITYPUB_HOST" default:"activitypub.awakari.com" required:"true"`
			Uri  string `envconfig:"API_ACTIVITYPUB_URI" default:"int-activitypub:50051" required:"true"`
		}
		Mastodon MastodonConfig
		Queue    QueueConfig
		Token    struct {
			Internal string `envconfig:"API_TOKEN_INTERNAL" required:"true"`
		}
	}
	Log struct {
		Level int `envconfig:"LOG_LEVEL" default:"-4" required:"true"`
	}
}

type MastodonConfig struct {
	Client struct {
		Tokens    []string `envconfig:"API_MASTODON_CLIENT_TOKENS" required:"true"`
		Hosts     []string `envconfig:"API_MASTODON_CLIENT_HOSTS" required:"true" default:"mastodon.social"`
		UserAgent string   `envconfig:"API_MASTODON_CLIENT_USER_AGENT" default:"awakari" required:"true"`
	}
	CountMin struct {
		Followers uint32 `envconfig:"API_MASTODON_COUNT_MIN_FOLLOWERS" default:"100" required:"true"`
		Posts     uint32 `envconfig:"API_MASTODON_COUNT_MIN_POSTS" default:"1000" required:"true"`
	}
	Endpoint struct {
		Protocol string `envconfig:"API_MASTODON_ENDPOINT_PROTOCOL" default:"https://" required:"true"`
		Accounts string `envconfig:"API_MASTODON_ENDPOINT_ACCOUNTS" default:"/api/v1/accounts" required:"true"`
		Search   string `envconfig:"API_MASTODON_ENDPOINT_SEARCH" default:"/api/v2/search" required:"true"`
	}
	Search struct {
		Limit uint32 `envconfig:"API_MASTODON_SEARCH_LIMIT" default:"10" required:"true"`
	}
}

type QueueConfig struct {
	BackoffError     time.Duration `envconfig:"API_QUEUE_BACKOFF_ERROR" default:"1s" required:"true"`
	Uri              string        `envconfig:"API_QUEUE_URI" default:"queue:50051" required:"true"`
	InterestsCreated struct {
		BatchSize uint32 `envconfig:"API_QUEUE_INTERESTS_CREATED_BATCH_SIZE" default:"1" required:"true"`
		Name      string `envconfig:"API_QUEUE_INTERESTS_CREATED_NAME" default:"int-mastodon" required:"true"`
		Subj      string `envconfig:"API_QUEUE_INTERESTS_CREATED_SUBJ" default:"interests-created" required:"true"`
	}
	InterestsUpdated struct {
		BatchSize uint32 `envconfig:"API_QUEUE_INTERESTS_UPDATED_BATCH_SIZE" default:"1" required:"true"`
		Name      string `envconfig:"API_QUEUE_INTERESTS_UPDATED_NAME" default:"int-mastodon" required:"true"`
		Subj      string `envconfig:"API_QUEUE_INTERESTS_UPDATED_SUBJ" default:"interests-updated" required:"true"`
	}
	SourceSse struct {
		BatchSize uint32 `envconfig:"API_QUEUE_SRC_SSE_BATCH_SIZE" default:"100" required:"true"`
		Name      string `envconfig:"API_QUEUE_SRC_SSE_NAME" default:"int-mastodon" required:"true"`
		Subj      string `envconfig:"API_QUEUE_SRC_SSE_SUBJ" default:"source-sse-mastodon" required:"true"`
	}
}

func NewConfigFromEnv() (cfg Config, err error) {
	err = envconfig.Process("", &cfg)
	return
}
