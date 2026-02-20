package content

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic/content"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

// 修改播客
func ModifyPodcastHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ModifyPodcastRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := content.NewModifyPodcastLogic(r.Context(), svcCtx)
		resp, err := l.ModifyPodcast(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
