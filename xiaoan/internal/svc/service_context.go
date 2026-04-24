package svc

import (
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
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
	Mysql          sqlx.SqlConn
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
	var mysql sqlx.SqlConn
	if c.MysqlConf.DataSource != "" {
		// 无 parseTime 时 datetime 会扫成 []uint8，go-zero/Scan 会失败（表现为「查询文章失败」等）
		mysql = sqlx.NewMysql(mysqlDataSourceWithParseTime(c.MysqlConf.DataSource))
	}
	return &ServiceContext{
		Config:         c,
		Mysql:          mysql,
		AuthRPC:        auth.NewAuthServiceClient(ac),
		QARPC:          qa.NewQAServiceClient(qc),
		ContentRPC:     content.NewContentServiceClient(cc),
		AuthMiddleware: middleware.NewAuthMiddleware(c.JWTConfig).Handle,
		IPMiddleware:   middleware.NewIPMiddleware(c.IP2RegionConfig).Handle,
	}
}

func mysqlDataSourceWithParseTime(dsn string) string {
	s := strings.TrimSpace(dsn)
	if s == "" {
		return s
	}
	lower := strings.ToLower(s)
	if strings.Contains(lower, "parsetime=true") {
		return s
	}
	if strings.Contains(s, "?") {
		return s + "&parseTime=true&loc=Local"
	}
	return s + "?parseTime=true&loc=Local"
}
