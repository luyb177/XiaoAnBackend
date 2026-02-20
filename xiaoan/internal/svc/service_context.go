package svc

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"

	auth "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	qa "github.com/luyb177/XiaoAnBackend/qa/pb/qa/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/config"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/middleware"
)

type ServiceContext struct {
	Config         config.Config
	AuthRPC        auth.AuthServiceClient
	QARPC          qa.QAServiceClient
	ContentRPC     content.ContentServiceClient
	AuthMiddleware rest.Middleware
	IPMiddleware   rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	ac := zrpc.MustNewClient(c.AuthRPC).Conn()
	qc := zrpc.MustNewClient(c.QARPC).Conn()
	cc := zrpc.MustNewClient(c.ContentRPC).Conn()
	return &ServiceContext{
		Config:         c,
		AuthRPC:        auth.NewAuthServiceClient(ac),
		QARPC:          qa.NewQAServiceClient(qc),
		ContentRPC:     content.NewContentServiceClient(cc),
		AuthMiddleware: middleware.NewAuthMiddleware(c.JWTConfig).Handle,
		IPMiddleware:   middleware.NewIPMiddleware(c.IP2RegionConfig).Handle,
	}
}
