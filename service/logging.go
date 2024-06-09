package service

import (
	"context"
	"fmt"
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

func (l logging) SearchAndAdd(ctx context.Context, subId, groupId, q string, limit uint32) (n uint32, err error) {
	n, err = l.svc.SearchAndAdd(ctx, subId, groupId, q, limit)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("service.SearchAndAdd(subId=%s, groupId=%s, q=%s): %d, %s", subId, groupId, q, n, err))
	return
}

func (l logging) ConsumeLiveStreamPublic(ctx context.Context) (err error) {
	l.log.Log(context.TODO(), logLevel(err), fmt.Sprintf("service.ConsumeLiveStreamPublic(): started"))
	err = l.svc.ConsumeLiveStreamPublic(ctx)
	l.log.Log(context.TODO(), logLevel(err), fmt.Sprintf("service.ConsumeLiveStreamPublic(): done, err=%s", err))
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