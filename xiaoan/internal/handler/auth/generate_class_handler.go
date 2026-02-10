package auth

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic/auth"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

// GenerateClassHandler 生成班级
func GenerateClassHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GenerateClassRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := auth.NewGenerateClassLogic(r.Context(), svcCtx)
		resp, err := l.GenerateClass(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
