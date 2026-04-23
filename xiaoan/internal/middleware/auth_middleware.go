package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/metadata"

	"github.com/luyb177/XiaoAnBackend/infra/jwt"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/config"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

type AuthMiddleware struct {
	r jwt.Handler
	logx.Logger
}

func NewAuthMiddleware(cfg config.JWTConfig) *AuthMiddleware {
	return &AuthMiddleware{
		r:      jwt.NewHandler(cfg.Secret, time.Duration(cfg.Expire)*time.Second),
		Logger: logx.WithContext(context.Background()),
	}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := parseAuthorizationToken(r.Header)
		if token == "" {
			httpx.OkJsonCtx(r.Context(), w, &types.Response{
				Code:    401,
				Message: "请先登录",
			})
			return
		}

		// todo 这里可以把 user 相关信息加密一下，然后解密
		claims, err := m.r.ParseJWTToken(token)
		if err != nil {
			m.Errorf("ParseJWTToken 解析token失败：err %v", err)
			httpx.OkJsonCtx(r.Context(), w, &types.Response{
				Code:    401,
				Message: "token解析失败",
			})
			return
		}

		role := claims.UserRole
		if role == "" {
			role = middleware.DefaultUserInfo().Role
		}

		ctx := r.Context()
		u := &middleware.UserInfo{
			UID:    claims.UserID,
			Role:   role,
			Status: claims.UserStatus,
		}
		ctx = middleware.WithUser(ctx, u)
		ctx = metadata.AppendToOutgoingContext(
			ctx,
			middleware.MdKeyUserID, strconv.FormatUint(claims.UserID, 10),
			middleware.MdKeyUserRole, role,
			middleware.MdKeyUserStatus, strconv.FormatInt(claims.UserStatus, 10),
		)

		next(w, r.WithContext(ctx))
	}
}

// parseAuthorizationToken 与常见前端/网关一致：支持「裸 JWT」或「Bearer <JWT>」
func parseAuthorizationToken(h http.Header) string {
	s := strings.TrimSpace(h.Get("Authorization"))
	if s == "" {
		return ""
	}
	const p = "Bearer "
	if len(s) > len(p) && strings.EqualFold(s[:len(p)], p) {
		return strings.TrimSpace(s[len(p):])
	}
	return s
}
