package middleware

import (
	"context"
	"fmt"

	"github.com/luyb177/XiaoAnBackend/auth/pkg/auth"

	"google.golang.org/grpc"
)

type ctxKey string

const (
	ctxKeyUser ctxKey = "user"
)

// 不需要鉴权的方法
//
//	/<proto包名>.<ServiceName>/<MethodName>
var noAuthMethods = map[string]struct{}{
	"/auth.AuthService/SendEmailCode":     {},
	"/auth.AuthService/ValidateEmailCode": {},
	"/auth.AuthService/Register":          {},
	"/auth.AuthService/Login":             {},
}

// UserUnaryInterceptor 用户服务拦截器
func UserUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 1. 跳过不需要鉴权的接口
	if _, ok := noAuthMethods[info.FullMethod]; ok {
		return handler(ctx, req)
	}

	// 2. 需要鉴权的接口，从 metadata 取用户信息
	uid, role, status, err := auth.GetUserFromMetadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("用户未登录或登录状态异常,%v", err)
	}

	// 3. 写入 context，供 logic 使用
	ctx = context.WithValue(ctx, ctxKeyUser, &UserInfo{
		UID:    uid,
		Role:   role,
		Status: status,
	})

	return handler(ctx, req)
}
