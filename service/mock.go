package service

import (
	"context"
	"github.com/awakari/int-mastodon/model"
)

type mock struct {
}

func NewServiceMock() Service {
	return mock{}
}

func (m mock) SearchAndAdd(ctx context.Context, subId, groupId, q string, limit uint32, typ model.SearchType) (n uint32, err error) {
	return 42, nil
}

func (m mock) HandleLiveStream(ctx context.Context) (err error) {
	return
}
