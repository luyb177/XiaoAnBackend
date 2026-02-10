package middleware

import (
	"context"

	"google.golang.org/grpc"

	"github.com/luyb177/XiaoAnBackend/content/pkg/ip2region"
)

const ctxKeyIPInfo ctxKey = "ip_info"

// 需要记录 IP 信息的方法
//
//	/<proto包名>.<ServiceName>/<MethodName>
var ipRecordMethods = map[string]struct{}{
	"/content.ContentService/AddComment": {},
}

// IPUnaryInterceptor IP 信息拦截器
func IPUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 1. 跳过不需要记录 IP 的接口
	if _, ok := ipRecordMethods[info.FullMethod]; !ok {
		return handler(ctx, req)
	}

	// 2. 需要记录 IP 的接口，从 metadata 取 IP 信息
	ipLocation := ip2region.GetIPFromMetadata(ctx)

	// 3. 写入 context，供 logic 使用
	if ipLocation != nil {
		ctx = context.WithValue(ctx, ctxKeyIPInfo, &IPInfo{
			ClientIP: ipLocation.ClientIP,
			Country:  ipLocation.Country,
			Province: ipLocation.Province,
			City:     ipLocation.City,
			ISP:      ipLocation.ISP,
			ISOCode:  ipLocation.ISOCode,
		})
	}
	return handler(ctx, req)
}
