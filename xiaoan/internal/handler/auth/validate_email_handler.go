package auth

import (
	"net/http"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic/auth"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// ValidateEmailHandler 验证邮箱验证码
func ValidateEmailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ValidateEmailRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := auth.NewValidateEmailLogic(r.Context(), svcCtx)
		resp, err := l.ValidateEmail(&req)
		if err != nil {
			logx.Errorf("ValidateEmailHandler error: %v", err)
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
