package service

import (
	"context"
	"fmt"
	"github.com/awakari/int-mastodon/model"
	"github.com/awakari/int-mastodon/util"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"log/slog"
)

type logging struct {
	svc Service
	log *slog.Logger
}

func NewServiceLogging(svc Service, log *slog.Logger) Service {
	return logging{
		svc: svc,
		log: log,
	}
}

func (l logging) SearchAndAdd(ctx context.Context, subId, groupId, q string, limit uint32, typ model.SearchType) (n uint32, err error) {
	n, err = l.svc.SearchAndAdd(ctx, subId, groupId, q, limit, typ)
	l.log.Log(ctx, util.LogLevel(err), fmt.Sprintf("service.SearchAndAdd(subId=%s, groupId=%s, q=%s, typ=%s): %d, %s", subId, groupId, q, typ.String(), n, err))
	return
}

func (l logging) HandleLiveStreamEvents(ctx context.Context, evts []*pb.CloudEvent) {
	l.svc.HandleLiveStreamEvents(ctx, evts)
	l.log.Debug(fmt.Sprintf("service.HandleLiveStreamEvents(%d)", len(evts)))
	return
}
