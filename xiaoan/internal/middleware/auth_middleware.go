package middleware

import (
	"context"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/config"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/pkg/ijwt"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	"time"
)

type AuthMiddleware struct {
	r ijwt.Handler
}

func NewAuthMiddleware(cfg config.JWTConfig) *AuthMiddleware {
	return &AuthMiddleware{
		r: ijwt.NewHandler(cfg.Secret, time.Duration(cfg.Expire)),
	}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			httpx.OkJsonCtx(r.Context(), w, &types.Response{
				Code:    401,
				Message: "请先登录",
			})
		}

		claims, err := m.r.ParseJWTToken(token)
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, &types.Response{
				Code:    401,
				Message: "token解析失败",
			})
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "user_id", claims.UserId)
		ctx = context.WithValue(ctx, "user_role", claims.UserRole)
		ctx = context.WithValue(ctx, "user_status", claims.UserStatus)
		next(w, r.WithContext(ctx))
	}
}
