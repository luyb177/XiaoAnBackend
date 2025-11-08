package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Email     EmailConfig
	RedisConf redis.RedisConf
	MysqlConf sqlx.SqlConf
}

type EmailConfig struct {
	From     string
	Password string
	SMTPHost string
	SMTPPort int
}
