package int_activitypub

import (
	"context"
	"errors"
	"fmt"
)

type Service interface {
	Create(ctx context.Context, addr, groupId, userId, subId, term string) (err error)
}

type service struct {
	client ServiceClient
}

var ErrInternal = errors.New("internal failure")

func NewService(client ServiceClient) Service {
	return service{
		client: client,
	}
}

func (svc service) Create(ctx context.Context, addr, groupId, userId, subId, term string) (err error) {
	_, err = svc.client.Create(ctx, &CreateRequest{
		Addr:    addr,
		GroupId: groupId,
		UserId:  userId,
	})
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrInternal, err)
	}
	return
}
