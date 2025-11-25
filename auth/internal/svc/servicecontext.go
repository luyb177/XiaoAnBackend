package svc

import (
	"github.com/luyb177/XiaoAnBackend/auth/internal/config"
	"github.com/luyb177/XiaoAnBackend/auth/internal/jwt"
	"github.com/luyb177/XiaoAnBackend/auth/internal/repo"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type ServiceContext struct {
	Config     config.Config
	RedisRepo  repo.RedisRepo
	Mysql      sqlx.SqlConn
	JWTHandler jwt.Handler
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		RedisRepo:  repo.NewRedisRepo(redis.MustNewRedis(c.RedisConf)),
		Mysql:      sqlx.NewMysql(c.MysqlConf.DataSource),
		JWTHandler: jwt.NewHandler(c.JWTConfig.Secret, time.Duration(c.JWTConfig.Expire)),
	}
}
