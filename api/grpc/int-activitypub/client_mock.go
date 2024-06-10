package int_activitypub

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type clientMock struct {
}

func NewClientMock() ServiceClient {
	return clientMock{}
}

func (cm clientMock) Create(ctx context.Context, req *CreateRequest, opts ...grpc.CallOption) (resp *CreateResponse, err error) {
	resp = &CreateResponse{
		Url: req.Addr,
	}
	switch req.Addr {
	case "fail":
		err = status.Error(codes.Internal, "internal failure")
	}
	return
}
