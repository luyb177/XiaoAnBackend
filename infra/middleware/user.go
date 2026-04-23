package middleware

import (
	"context"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/luyb177/XiaoAnBackend/infra/constants"
)

type ctxKey string

const (
	ctxKeyUser ctxKey = "user"
)

const (
	MdKeyUserID     = "user_id"
	MdKeyUserRole   = "user_role"
	MdKeyUserStatus = "user_status"
)

// 不需要鉴权的方法
//
//	/<proto包名>.<ServiceName>/<MethodName>
var noAuthMethods = map[string]struct{}{}

// UserUnaryInterceptor 用户服务拦截器
func UserUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 1. 跳过不需要鉴权的接口
	if _, ok := noAuthMethods[info.FullMethod]; ok {
		return handler(ctx, req)
	}

	// 2. 需要鉴权的接口，从 metadata 取用户信息
	userInfo := extractUserFromMetadata(ctx)

	// 3. 写入 context，供 logic 使用
	ctx = context.WithValue(ctx, ctxKeyUser, userInfo)

	return handler(ctx, req)
}

type UserInfo struct {
	UID    uint64
	Role   string
	Status int64
}

func GetUser(ctx context.Context) (*UserInfo, bool) {
	if u, ok := ctx.Value(ctxKeyUser).(*UserInfo); ok && u != nil {
		return u, true
	}
	// HTTP 网关在多轮 context 包一层时，*UserInfo 可能不在 Value 链上；鉴权后写入的
	// gRPC Outgoing metadata（user_id 等）仍在，可据此恢复
	if u, ok := userFromOutgoingMetadata(ctx); ok {
		return u, true
	}
	return nil, false
}

func userFromOutgoingMetadata(ctx context.Context) (*UserInfo, bool) {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok || len(md) == 0 {
		return nil, false
	}
	// FromOutgoingContext 的键名均为小写，与 MdKey* 一致
	get := func(key string) string {
		vals := md[key]
		if len(vals) == 0 {
			return ""
		}
		return vals[0]
	}
	uidStr := get(MdKeyUserID)
	if uidStr == "" {
		return nil, false
	}
	uid, _ := strconv.ParseUint(uidStr, 10, 64)
	status, _ := strconv.ParseInt(get(MdKeyUserStatus), 10, 64)
	role := get(MdKeyUserRole)
	if role == "" {
		role = constants.GUEST
	}
	return &UserInfo{UID: uid, Role: role, Status: status}, true
}

// WithUser 将已认证用户信息写入 context，供网关在 HTTP 链路上使用（与 gRPC 拦截器写入的 ctx 一致）
func WithUser(ctx context.Context, u *UserInfo) context.Context {
	if u == nil {
		return ctx
	}
	return context.WithValue(ctx, ctxKeyUser, u)
}

func DefaultUserInfo() *UserInfo {
	return &UserInfo{
		UID:    0,
		Role:   constants.GUEST,
		Status: 1,
	}
}

func extractUserFromMetadata(ctx context.Context) *UserInfo {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return DefaultUserInfo()
	}

	get := func(key string) string {
		values := md.Get(key)
		if len(values) == 0 {
			return ""
		}
		return values[0]
	}

	uid, _ := strconv.ParseUint(get(MdKeyUserID), 10, 64)
	status, _ := strconv.ParseInt(get(MdKeyUserStatus), 10, 64)

	role := get(MdKeyUserRole)
	if role == "" {
		role = constants.GUEST
	}

	return &UserInfo{
		UID:    uid,
		Role:   role,
		Status: status,
	}
}
