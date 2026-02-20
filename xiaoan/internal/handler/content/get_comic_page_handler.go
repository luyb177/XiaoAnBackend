package content

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic/content"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

// 获取漫画页面
func GetComicPageHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetComicPageRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := content.NewGetComicPageLogic(r.Context(), svcCtx)
		resp, err := l.GetComicPage(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
