package queue

import (
	"os"

	"github.com/hibiken/asynq"
)

var AsynqClient *asynq.Client

func InitClient() {
	AsynqClient = asynq.NewClient(asynq.RedisClientOpt{Addr: os.Getenv("REDIS_URL")})
}
