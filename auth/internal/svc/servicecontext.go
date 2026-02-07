package svc

import (
	"time"

	"github.com/luyb177/XiaoAnBackend/auth/internal/config"
	"github.com/luyb177/XiaoAnBackend/auth/internal/jwt"
	"github.com/luyb177/XiaoAnBackend/auth/internal/repo/email"
	"github.com/luyb177/XiaoAnBackend/auth/internal/repo/redisqueue"
	"github.com/luyb177/XiaoAnBackend/auth/pkg/taskqueue"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config     config.Config
	RedisRepo  RedisRepo
	Mysql      sqlx.SqlConn
	TaskQueue  taskqueue.TaskQueue
	JWTHandler jwt.Handler
}

func NewServiceContext(c config.Config) *ServiceContext {
	keys := taskqueue.QueueKey{
		Pending:    "auth:pending",
		Processing: "auth:processing",
		Retry:      "auth:retry",
		DLQ:        "auth:dlq",
	}

	rds := redis.MustNewRedis(c.RedisConf)

	rr := RedisRepo{
		EmailRepo: email.NewRedisRepo(rds),
	}
	tq := redisqueue.NewRedisTaskQueue(rds, keys)

	return &ServiceContext{
		Config:     c,
		RedisRepo:  rr,
		Mysql:      sqlx.NewMysql(c.MysqlConf.DataSource),
		JWTHandler: jwt.NewHandler(c.JWTConfig.Secret, time.Duration(c.JWTConfig.Expire)),
		TaskQueue:  tq,
	}
}

type RedisRepo struct {
	EmailRepo email.Repository
}
