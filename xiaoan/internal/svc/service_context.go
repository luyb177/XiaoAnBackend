package svc

import (
	auth "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	qa "github.com/luyb177/XiaoAnBackend/qa/pb/qa/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config  config.Config
	AuthRpc auth.AuthServiceClient
	QARpc   qa.QAServiceClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	cc := zrpc.MustNewClient(c.AuthRpc).Conn()
	qc := zrpc.MustNewClient(c.QARpc).Conn()
	return &ServiceContext{
		Config:  c,
		AuthRpc: auth.NewAuthServiceClient(cc),
		QARpc:   qa.NewQAServiceClient(qc),
	}
}
