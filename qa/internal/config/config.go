package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	LLMClientConfig LLMClientConfig
	MysqlConf       sqlx.SqlConf
	RedisConf       redis.RedisConf
}

type LLMClientConfig struct {
	Model   string
	BaseURL string
	APIKey  string
}
