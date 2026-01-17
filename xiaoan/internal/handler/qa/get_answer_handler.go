package qa

import (
	"net/http"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic/qa"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetAnswerHandler 获取答案
func GetAnswerHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetAnswerRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := qa.NewGetAnswerLogic(r.Context(), svcCtx)
		resp, err := l.GetAnswer(&req)
		if err != nil {
			logx.Errorf("GetAnswerHandler error: %v", err)
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
