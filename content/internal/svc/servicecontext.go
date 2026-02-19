package svc

import (
	"github.com/luyb177/XiaoAnBackend/infra/queue"
	"github.com/luyb177/XiaoAnBackend/infra/queue/redisqueue"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/luyb177/XiaoAnBackend/content/internal/config"
	"github.com/luyb177/XiaoAnBackend/content/internal/repo/collect"
	"github.com/luyb177/XiaoAnBackend/content/internal/repo/like"
)

type ServiceContext struct {
	Config      config.Config
	Mysql       sqlx.SqlConn
	TaskQueue   queue.TaskQueue
	LikeRepo    like.Repository
	CollectRepo collect.Repository
}

func NewServiceContext(c config.Config) *ServiceContext {
	keys := queue.QueueKey{
		Pending:    "content:pending",
		Processing: "content:processing",
		Retry:      "content:retry",
		DLQ:        "content:dlq",
	}
	rds := redis.MustNewRedis(c.RedisConf)

	tq := redisqueue.NewRedisTaskQueue(rds, keys)
	lr := like.NewRepository(rds)
	cr := collect.NewRepository(rds)

	return &ServiceContext{
		Config:      c,
		Mysql:       sqlx.NewMysql(c.MysqlConf.DataSource),
		TaskQueue:   tq,
		LikeRepo:    lr,
		CollectRepo: cr,
	}
}
