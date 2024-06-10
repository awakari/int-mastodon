package int_activitypub

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_Create(t *testing.T) {
	svc := NewService(NewClientMock())
	cases := map[string]struct {
		addr    string
		groupId string
		userId  string
		subId   string
		term    string
		err     error
	}{
		"ok": {
			addr: "addr1",
		},
		"fail": {
			addr: "fail",
			err:  ErrInternal,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err := svc.Create(context.TODO(), c.addr, c.groupId, c.userId, c.subId, c.term)
			assert.ErrorIs(t, err, c.err)
		})
	}
}
