package grpc

import (
	"context"
	"fmt"
	"github.com/awakari/int-mastodon/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"os"
	"testing"
)

var port uint16 = 50051

var log = slog.Default()

func TestMain(m *testing.M) {
	svc := service.NewServiceMock()
	svc = service.NewServiceLogging(svc, log)
	go func() {
		err := Serve(port, svc)
		if err != nil {
			log.Error("", err)
		}
	}()
	code := m.Run()
	os.Exit(code)
}

func TestServiceClient_SearchAndAdd(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		req *SearchAndAddRequest
		n   uint32
		err error
	}{
		"ok": {
			req: &SearchAndAddRequest{
				Q: "ok",
			},
			n: 42,
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			_, err := client.SearchAndAdd(context.TODO(), c.req)
			assert.ErrorIs(t, err, c.err)
		})
	}
}
