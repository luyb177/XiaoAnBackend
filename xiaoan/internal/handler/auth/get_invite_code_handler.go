package auth

import (
	"net/http"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic/auth"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取邀请码
func GetInviteCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetInviteCodeRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := auth.NewGetInviteCodeLogic(r.Context(), svcCtx)
		resp, err := l.GetInviteCode(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
