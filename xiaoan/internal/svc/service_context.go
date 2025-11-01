package svc

import (
	pb "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config  config.Config
	AuthRpc pb.AuthServiceClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	cc := zrpc.MustNewClient(c.AuthRpc).Conn()
	return &ServiceContext{
		Config:  c,
		AuthRpc: pb.NewAuthServiceClient(cc),
	}
}
