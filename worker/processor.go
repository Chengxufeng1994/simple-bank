package worker

import (
	"context"

	"github.com/go-redis/redis/v8"

	db "github.com/Chengxufeng1994/simple-bank/db/sqlc"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

type TaskProcessor interface {
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
	Start() error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
}

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store) TaskProcessor {
	logger := NewLogger()
	redis.SetLogger(logger)
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().Err(err).Str("type", task.Type()).
					Bytes("payload", task.Payload()).Msg("process task failed")
			}),
			Logger: logger,
		},
	)

	return &RedisTaskProcessor{
		server: server,
		store:  store,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TypeSendVerifyEmail, processor.ProcessTaskSendVerifyEmail)

	return processor.server.Run(mux)
}
