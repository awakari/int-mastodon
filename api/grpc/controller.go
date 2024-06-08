package grpc

import (
	"context"
	"github.com/awakari/int-mastodon/service"
)

type controller struct {
	search service.Service
}

func NewController(search service.Service) ServiceServer {
	return controller{
		search: search,
	}
}

func (c controller) SearchAndAdd(ctx context.Context, req *SearchAndAddRequest) (resp *SearchAndAddResponse, err error) {
	resp = &SearchAndAddResponse{}
	resp.N, err = c.search.SearchAndAdd(ctx, req.SubId, req.GroupId, req.Q, req.Limit)
	return
}
