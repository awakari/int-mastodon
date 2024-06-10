package main

import (
	"context"
	"fmt"
	"github.com/awakari/client-sdk-go/api"
	apiGrpc "github.com/awakari/int-mastodon/api/grpc"
	apiGrpcAp "github.com/awakari/int-mastodon/api/grpc/int-activitypub"
	"github.com/awakari/int-mastodon/config"
	"github.com/awakari/int-mastodon/service"
	"github.com/awakari/int-mastodon/service/writer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	//
	cfg, err := config.NewConfigFromEnv()
	if err != nil {
		panic(fmt.Sprintf("failed to load the config from env: %s", err))
	}
	//
	opts := slog.HandlerOptions{
		Level: slog.Level(cfg.Log.Level),
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	log.Info("starting the update for the feeds")
	//
	var clientAwk api.Client
	clientAwk, err = api.
		NewClientBuilder().
		WriterUri(cfg.Api.Writer.Uri).
		Build()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize the Awakari API client: %s", err))
	}
	defer clientAwk.Close()
	log.Info("initialized the Awakari API client")
	//
	connAp, err := grpc.NewClient(cfg.Api.ActivityPub.Uri, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	log.Info("connected to the int-activitypub service")
	clientAp := apiGrpcAp.NewServiceClient(connAp)
	svcActivityPub := apiGrpcAp.NewService(clientAp)
	svcActivityPub = apiGrpcAp.NewServiceLogging(svcActivityPub, log)
	//
	svcWriter := writer.NewService(clientAwk, cfg.Api.Writer.Backoff)
	svcWriter = writer.NewLogging(svcWriter, log)
	//
	clientHttp := &http.Client{}
	svc := service.NewService(clientHttp, cfg.Api.Mastodon.Client.UserAgent, cfg.Api.Mastodon, svcActivityPub, svcWriter)
	svc = service.NewServiceLogging(svc, log)
	//
	go func() {
		for {
			err = svc.HandleLiveStream(context.Background())
			if err != nil {
				panic(err)
			}
		}
	}()
	//
	log.Info(fmt.Sprintf("starting to listen the gRPC API @ port #%d...", cfg.Api.Port))
	if err = apiGrpc.Serve(cfg.Api.Port, svc); err != nil {
		panic(err)
	}
}
