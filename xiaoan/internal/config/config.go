package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	AuthRpc zrpc.RpcClientConf
	QARpc   zrpc.RpcClientConf
	//ContentRpc zrpc.RpcClientConf
	JWTConfig JWTConfig
}

type JWTConfig struct {
	Secret string
	Expire int64
}
