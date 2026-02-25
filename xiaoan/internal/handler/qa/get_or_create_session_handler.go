// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package qa

import (
	"net/http"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic/qa"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetOrCreateSessionHandler 获取或创建问答会话
func GetOrCreateSessionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetOrCreateSessionRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := qa.NewGetOrCreateSessionLogic(r.Context(), svcCtx)
		resp, err := l.GetOrCreateSession(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
