package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	AuthRpc         zrpc.RpcClientConf
	QARpc           zrpc.RpcClientConf
	ContentRpc      zrpc.RpcClientConf
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
