package main

import (
	"context"
	"fmt"
	"github.com/awakari/client-sdk-go/api"
	apiGrpc "github.com/awakari/int-mastodon/api/grpc"
	apiGrpcAp "github.com/awakari/int-mastodon/api/grpc/int-activitypub"
	"github.com/awakari/int-mastodon/api/grpc/queue"
	"github.com/awakari/int-mastodon/config"
	"github.com/awakari/int-mastodon/model"
	"github.com/awakari/int-mastodon/service"
	"github.com/awakari/int-mastodon/service/writer"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

const ceKeyGroupId = "awakarigroupid"
const ceKeyQueriesCompl = "queriescompl"
const ceKeyPublic = "public"

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
	svcWriter := writer.NewService(clientAwk, cfg.Api.Writer.Backoff, log)
	svcWriter = writer.NewLogging(svcWriter, log)
	//
	clientHttp := &http.Client{}
	svc := service.NewService(clientHttp, cfg.Api.Mastodon.Client.UserAgent, cfg.Api.Mastodon, svcActivityPub, svcWriter, cfg.Api.Event.Type)
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

	// init queues
	connQueue, err := grpc.NewClient(cfg.Api.Queue.Uri, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	log.Info("connected to the queue service")
	clientQueue := queue.NewServiceClient(connQueue)
	svcQueue := queue.NewService(clientQueue)
	svcQueue = queue.NewLoggingMiddleware(svcQueue, log)
	err = svcQueue.SetConsumer(context.TODO(), cfg.Api.Queue.InterestsCreated.Name, cfg.Api.Queue.InterestsCreated.Subj)
	if err != nil {
		panic(err)
	}
	log.Info(fmt.Sprintf("initialized the %s queue", cfg.Api.Queue.InterestsCreated.Name))
	go func() {
		err = consumeQueue(
			context.Background(),
			svc,
			svcQueue,
			cfg.Api.Queue.InterestsCreated.Name,
			cfg.Api.Queue.InterestsCreated.Subj,
			cfg.Api.Queue.InterestsCreated.BatchSize,
			cfg,
			log,
		)
		if err != nil {
			panic(err)
		}
	}()
	err = svcQueue.SetConsumer(context.TODO(), cfg.Api.Queue.InterestsUpdated.Name, cfg.Api.Queue.InterestsUpdated.Subj)
	if err != nil {
		panic(err)
	}
	log.Info(fmt.Sprintf("initialized the %s queue", cfg.Api.Queue.InterestsUpdated.Name))
	go func() {
		err = consumeQueue(
			context.Background(),
			svc,
			svcQueue,
			cfg.Api.Queue.InterestsUpdated.Name,
			cfg.Api.Queue.InterestsUpdated.Subj,
			cfg.Api.Queue.InterestsUpdated.BatchSize,
			cfg,
			log,
		)
		if err != nil {
			panic(err)
		}
	}()

	log.Info(fmt.Sprintf("starting to listen the gRPC API @ port #%d...", cfg.Api.Port))
	err = apiGrpc.Serve(cfg.Api.Port, svc)
	if err != nil {
		panic(err)
	}
}

func consumeQueue(
	ctx context.Context,
	svc service.Service,
	svcQueue queue.Service,
	name, subj string,
	batchSize uint32,
	cfg config.Config,
	log *slog.Logger,
) (err error) {
	for {
		err = svcQueue.ReceiveMessages(ctx, name, subj, batchSize, func(evts []*pb.CloudEvent) (err error) {
			consumeEvents(ctx, svc, evts, cfg, log)
			return
		})
		if err != nil {
			panic(err)
		}
	}
}

func consumeEvents(
	ctx context.Context,
	svc service.Service,
	evts []*pb.CloudEvent,
	cfg config.Config,
	log *slog.Logger,
) {
	log.Debug(fmt.Sprintf("consumeEvents(%d))\n", len(evts)))
	for _, evt := range evts {

		interestId := evt.GetTextData()
		var groupId string
		if groupIdAttr, groupIdIdPresent := evt.Attributes[ceKeyGroupId]; groupIdIdPresent {
			groupId = groupIdAttr.GetCeString()
		}
		if groupId == "" {
			log.Error(fmt.Sprintf("interest %s event: empty group id, skipping", interestId))
			continue
		}

		publicAttr, publicAttrPresent := evt.Attributes[ceKeyPublic]
		switch publicAttrPresent && publicAttr.GetCeBoolean() {
		case true:
			actor := interestId + "@" + cfg.Api.ActivityPub.Host
			_, _ = svc.SearchAndAdd(ctx, interestId, groupId, actor, 1, model.SearchTypeAccounts)
		default:
			log.Debug(fmt.Sprintf("interest %s event: public: %t/%t", interestId, publicAttrPresent, publicAttr.GetCeBoolean()))
		}

		var queries []string
		if queriesComplAttr, queriesComplPresent := evt.Attributes[ceKeyQueriesCompl]; queriesComplPresent {
			queries = strings.Split(queriesComplAttr.GetCeString(), "\n")
		}
		switch len(queries) {
		case 0:
			log.Debug(fmt.Sprintf("interest %s event: no queries, skipping the sources discovery", interestId))
		default:
			for _, q := range queries {
				_, _ = svc.SearchAndAdd(ctx, interestId, groupId, q, cfg.Api.Mastodon.Search.Limit, model.SearchTypeStatuses)
			}
		}
	}
	return
}
