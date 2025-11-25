package svc

import (
	"github.com/luyb177/XiaoAnBackend/content/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config      config.Config
	MinioClient *minio.Client
	Mysql       sqlx.SqlConn
}

func NewServiceContext(c config.Config) *ServiceContext {
	minioClient, err := minio.New(c.MinioConf.EndPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.MinioConf.AccessKeyID, c.MinioConf.SecretAccessKey, ""),
		Secure: c.MinioConf.UseSSL,
	})
	if err != nil {
		logx.Errorf("minio new error: %v", err)
	}
	return &ServiceContext{
		Config:      c,
		MinioClient: minioClient,
		Mysql:       sqlx.NewMysql(c.MysqlConf.DataSource),
	}
}
