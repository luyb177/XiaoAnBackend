package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/luyb177/XiaoAnBackend/content/internal/config"
	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/server"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/internal/worker"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"log"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/content.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		v1.RegisterContentServiceServer(grpcServer, server.NewContentServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	s.AddUnaryInterceptors(middleware.UserUnaryInterceptor)

	//启动 worker
	w := worker.NewWorker(ctx)
	go func() {
		if err := w.Start(context.Background()); err != nil {
			log.Fatalf("worker start error: %v", err)
		}
	}()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
