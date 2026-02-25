package svc

import (
	"github.com/luyb177/XiaoAnBackend/infra/queue"
	"github.com/luyb177/XiaoAnBackend/infra/queue/redisqueue"
	"github.com/luyb177/XiaoAnBackend/qa/internal/config"
	"github.com/luyb177/XiaoAnBackend/qa/internal/repo/chatmessage"
	"github.com/luyb177/XiaoAnBackend/qa/internal/repo/chatsession"
	llmopenai "github.com/luyb177/XiaoAnBackend/qa/pkg/llm/openai"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config          config.Config
	LLMClient       llmopenai.LLMClient
	Mysql           sqlx.SqlConn
	TaskQueue       queue.TaskQueue
	ChatSessionRepo chatsession.Repository
	ChatMessageRepo chatmessage.Repository
}

func NewServiceContext(c config.Config) *ServiceContext {
	lc := llmopenai.NewLLMClient(c.LLMClientConfig)

	keys := queue.QueueKey{
		Pending:    "qa:pending",
		Processing: "qa:processing",
		Retry:      "qa:retry",
		DLQ:        "qa:dlq",
	}

	rds := redis.MustNewRedis(c.RedisConf)

	tq := redisqueue.NewRedisTaskQueue(rds, keys)
	csr := chatsession.NewRepository(rds)
	cmr := chatmessage.NewRepository(rds)

	return &ServiceContext{
		Config:          c,
		LLMClient:       lc,
		Mysql:           sqlx.NewMysql(c.MysqlConf.DataSource),
		TaskQueue:       tq,
		ChatSessionRepo: csr,
		ChatMessageRepo: cmr,
	}
}
