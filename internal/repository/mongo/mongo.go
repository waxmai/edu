package mongo

import (
	"context"
	"log"
	"sync"
	"time"

	"edu-schedule-system/configs"
	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/pkg/timeutil"
	"edu-schedule-system/internal/pkg/trace"

	"go.mongodb.org/mongo-driver/event"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client *mongoDriver.Client
	once   sync.Once
)

func GetMongoClient() *mongoDriver.Client {
	mongoInfo := new(trace.Mongo)

	loggingMonitor := &event.CommandMonitor{
		Started: func(ctx context.Context, startedEvent *event.CommandStartedEvent) {
			stdCtx, ok := ctx.(core.StdContext)
			if !ok {
				log.Println("mongo monitor started ctx illegal")
				return
			}

			if stdCtx.Trace != nil {
				mongoInfo = new(trace.Mongo)
				mongoInfo.Time = timeutil.CSTLayoutString()
				mongoInfo.Database = startedEvent.DatabaseName
				mongoInfo.Command = startedEvent.Command.String()
			}
		},
		Succeeded: func(ctx context.Context, succeededEvent *event.CommandSucceededEvent) {
			stdCtx, ok := ctx.(core.StdContext)
			if !ok {
				log.Println("mongo monitor succeeded ctx illegal")
				return
			}

			if stdCtx.Trace != nil {
				mongoInfo.Reply = succeededEvent.Reply.String()
				mongoInfo.CostSeconds = float64(succeededEvent.DurationNanos) / 1e9

				stdCtx.Trace.AppendMongo(mongoInfo)
			}
		},
	}

	once.Do(func() {
		cfg := configs.Get().Mongo

		clientOptions := options.Client().ApplyURI(cfg.URI).SetConnectTimeout(10 * time.Second)

		if cfg.UserName != "" || cfg.Password != "" || cfg.AuthSource != "" {
			clientOptions.SetAuth(options.Credential{
				Username:   cfg.UserName,
				Password:   cfg.Password,
				AuthSource: cfg.AuthSource,
			})
		}

		clientOptions.SetMonitor(loggingMonitor)

		var err error
		client, err = mongoDriver.Connect(context.Background(), clientOptions)
		if err != nil {
			panic(err)
		}
	})

	return client
}
