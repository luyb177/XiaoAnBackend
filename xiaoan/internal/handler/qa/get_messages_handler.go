// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package qa

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic/qa"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

// 获取问答消息列表
func GetMessagesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetMessagesRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := qa.NewGetMessagesLogic(r.Context(), svcCtx)
		resp, err := l.GetMessages(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
