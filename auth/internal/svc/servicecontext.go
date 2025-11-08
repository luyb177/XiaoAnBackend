package svc

import (
	"github.com/luyb177/XiaoAnBackend/auth/internal/config"
	"github.com/luyb177/XiaoAnBackend/auth/internal/repo"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type ServiceContext struct {
	Config    config.Config
	RedisRepo repo.RedisRepo
	Mysql     sqlx.SqlConn
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:    c,
		RedisRepo: repo.NewRedisRepo(redis.MustNewRedis(c.RedisConf)),
		Mysql:     sqlx.NewMysql(c.MysqlConf.DataSource),
	}
}
