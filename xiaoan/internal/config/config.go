package config

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	// Mysql 与 content 服务使用同一库，用于只读查询（如用户作品列表）
	MysqlConf       sqlx.SqlConf
	AuthRPC         zrpc.RpcClientConf
	QARPC           zrpc.RpcClientConf
	ContentRPC      zrpc.RpcClientConf
	JWTConfig       JWTConfig
	IP2RegionConfig IP2RegionConfig
}

type JWTConfig struct {
	Secret string
	Expire int64
}

type IP2RegionConfig struct {
	V4 string
	V6 string
}
