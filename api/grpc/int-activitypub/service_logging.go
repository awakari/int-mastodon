package int_activitypub

import (
	"context"
	"fmt"
	"log/slog"
)

type svcLogging struct {
	svc Service
	log *slog.Logger
}

func NewServiceLogging(svc Service, log *slog.Logger) Service {
	return svcLogging{
		svc: svc,
		log: log,
	}
}

func (sl svcLogging) Create(ctx context.Context, addr, groupId, userId, subId, term string) (err error) {
	err = sl.svc.Create(ctx, addr, groupId, userId, subId, term)
	sl.log.Log(ctx, sl.logLevel(err), fmt.Sprintf("int-activitypub.Create(%s, %s, %s, %s, %s): %s", addr, groupId, userId, subId, term, err))
	return
}

func (sl svcLogging) logLevel(err error) (lvl slog.Level) {
	switch err {
	case nil:
		lvl = slog.LevelInfo
	default:
		lvl = slog.LevelError
	}
	return
}
