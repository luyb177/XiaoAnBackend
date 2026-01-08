package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	MinioConf MinioConf
	MysqlConf sqlx.SqlConf
	RedisConf redis.RedisConf
}

type MinioConf struct {
	EndPoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	ContentBucket   string
}
