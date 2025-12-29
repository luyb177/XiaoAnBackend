package svc

import (
	auth "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	qa "github.com/luyb177/XiaoAnBackend/qa/pb/qa/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/config"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/middleware"
	"github.com/zeromicro/go-zero/rest"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	AuthRpc        auth.AuthServiceClient
	QARpc          qa.QAServiceClient
	ContentRpc     content.ContentServiceClient
	AuthMiddleware rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	ac := zrpc.MustNewClient(c.AuthRpc).Conn()
	qc := zrpc.MustNewClient(c.QARpc).Conn()
	cc := zrpc.MustNewClient(c.ContentRpc).Conn()
	return &ServiceContext{
		Config:         c,
		AuthRpc:        auth.NewAuthServiceClient(ac),
		QARpc:          qa.NewQAServiceClient(qc),
		ContentRpc:     content.NewContentServiceClient(cc),
		AuthMiddleware: middleware.NewAuthMiddleware(c.JWTConfig).Handle,
	}
}
