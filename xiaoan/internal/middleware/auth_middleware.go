package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/config"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/pkg/ijwt"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/metadata"
)

type AuthMiddleware struct {
	r ijwt.Handler
	logx.Logger
}

func NewAuthMiddleware(cfg config.JWTConfig) *AuthMiddleware {
	return &AuthMiddleware{
		r:      ijwt.NewHandler(cfg.Secret, time.Duration(cfg.Expire)),
		Logger: logx.WithContext(context.Background()),
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
			return
		}

		claims, err := m.r.ParseJWTToken(token)
		if err != nil {
			m.Logger.Errorf("ParseJWTToken 解析token失败：err %v", err)
			httpx.OkJsonCtx(r.Context(), w, &types.Response{
				Code:    401,
				Message: "token解析失败",
			})
			return
		}

		md := metadata.New(map[string]string{
			"user_id":     strconv.FormatUint(claims.UserId, 10),
			"user_role":   claims.UserRole,
			"user_status": strconv.FormatUint(uint64(claims.UserStatus), 10),
		})
		ctx := metadata.NewOutgoingContext(r.Context(), md)
		next(w, r.WithContext(ctx))
	}
}
