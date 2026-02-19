package middleware

import (
	"context"

	"github.com/luyb177/XiaoAnBackend/infra/jwt"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"

	"net/http"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/metadata"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/config"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

type AuthMiddleware struct {
	r jwt.Handler
	logx.Logger
}

func NewAuthMiddleware(cfg config.JWTConfig) *AuthMiddleware {
	return &AuthMiddleware{
		r:      jwt.NewHandler(cfg.Secret, time.Duration(cfg.Expire)),
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

		ctx := r.Context()
		ctx = metadata.AppendToOutgoingContext(
			ctx,
			middleware.MdKeyUserID, strconv.FormatUint(claims.UserID, 10),
			middleware.MdKeyUserRole, claims.UserRole,
			middleware.MdKeyUserStatus, strconv.FormatUint(uint64(claims.UserStatus), 10),
		)

		next(w, r.WithContext(ctx))
	}
}
