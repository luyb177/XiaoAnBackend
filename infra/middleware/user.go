package middleware

import (
	"context"
	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
	user, ok := ctx.Value(ctxKeyUser).(*UserInfo)
	return user, ok
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
