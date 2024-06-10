package int_activitypub

import "context"

type serviceMock struct {
}

func NewServiceMock() Service {
	return serviceMock{}
}

func (sm serviceMock) Create(ctx context.Context, addr, groupId, userId, subId, term string) (err error) {
	switch addr {
	case "fail":
		err = ErrInternal
	}
	return
}
