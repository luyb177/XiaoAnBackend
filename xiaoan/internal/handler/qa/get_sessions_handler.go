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

// 获取问答会话列表
func GetSessionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetSessionsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := qa.NewGetSessionsLogic(r.Context(), svcCtx)
		resp, err := l.GetSessions(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
