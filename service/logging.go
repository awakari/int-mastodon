package service

import (
	"context"
	"fmt"
	"github.com/awakari/int-mastodon/model"
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
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("service.SearchAndAdd(subId=%s, groupId=%s, q=%s, typ=%s): %d, %s", subId, groupId, q, typ.String(), n, err))
	return
}

func (l logging) HandleLiveStream(ctx context.Context) (err error) {
	l.log.Log(context.TODO(), logLevel(err), fmt.Sprintf("service.HandleLiveStream(): started"))
	err = l.svc.HandleLiveStream(ctx)
	l.log.Log(context.TODO(), logLevel(err), fmt.Sprintf("service.HandleLiveStream(): done, err=%s", err))
	return
}

func logLevel(err error) (lvl slog.Level) {
	switch err {
	case nil:
		lvl = slog.LevelDebug
	default:
		lvl = slog.LevelError
	}
	return
}
