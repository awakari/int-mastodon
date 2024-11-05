package service

import (
	"context"
	"github.com/awakari/int-mastodon/model"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
)

type mock struct {
}

func NewServiceMock() Service {
	return mock{}
}

func (m mock) SearchAndAdd(ctx context.Context, subId, groupId, q string, limit uint32, typ model.SearchType) (n uint32, err error) {
	return 42, nil
}

func (m mock) HandleLiveStreamEvents(ctx context.Context, evts []*pb.CloudEvent) {
	return
}
