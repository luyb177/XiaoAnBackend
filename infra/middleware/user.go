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

func UserStreamInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// 1. 跳过不需要鉴权的方法
	if _, ok := noAuthMethods[info.FullMethod]; ok {
		return handler(srv, ss)
	}

	// 2. 从 metadata 提取用户
	userInfo := extractUserFromMetadata(ss.Context())

	// 3. 写入新的 context
	newCtx := context.WithValue(ss.Context(), ctxKeyUser, userInfo)

	// 4. 包装 ServerStream
	// 这里包装一层 ServerStream，重写 Context() 方法，使其返回新的 context
	// 这样在 handler 中调用 ss.Context() 就能拿到新的 context，从而获取用户信息
	// 因为 gRPC 的 stream context 是只读的
	// 不能直接修改 ss.Context()，只能通过包装 ServerStream 来实现
	wrapped := &userServerStream{
		ServerStream: ss,
		ctx:          newCtx,
	}

	// 5. 继续执行 handler
	return handler(srv, wrapped)
}

type userServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *userServerStream) Context() context.Context {
	return s.ctx
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
