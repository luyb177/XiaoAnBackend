package middleware

import (
	"context"
	"fmt"
	"github.com/luyb177/XiaoAnBackend/auth/utils"

	"google.golang.org/grpc"
)

const (
	ctxKeyUserID     string = "user_id"
	ctxKeyUserRole   string = "user_role"
	ctxKeyUserStatus string = "user_status"
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
	uid, role, status, err := utils.GetUserFromMetadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("用户未登录或登录状态异常,%v", err)
	}

	// 3. 写入 context，供 logic 使用
	ctx = context.WithValue(ctx, ctxKeyUserID, uid)
	ctx = context.WithValue(ctx, ctxKeyUserRole, role)
	ctx = context.WithValue(ctx, ctxKeyUserStatus, status)

	return handler(ctx, req)
}
