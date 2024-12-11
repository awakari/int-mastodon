package int_activitypub

import (
	"context"
	"fmt"
	"github.com/awakari/int-mastodon/util"
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
	sl.log.Log(ctx, util.LogLevel(err), fmt.Sprintf("int-activitypub.Create(%s, %s, %s, %s, %s): %s", addr, groupId, userId, subId, term, err))
	return
}
