// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package qa

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/threading"
	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic/qa"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

// AskHandler 流式问答接口
func AskHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AskRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// Buffer size of 16 is chosen as a reasonable default to balance throughput and memory usage.
		// You can change this based on your application's needs.
		// if your go-zero version less than 1.8.1, you need to add 3 lines below.
		// w.Header().Set("Content-Type", "text/event-stream")
		// w.Header().Set("Cache-Control", "no-cache")
		// w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		client := make(chan *types.AskStreamReply, 16)

		l := qa.NewAskLogic(r.Context(), svcCtx)
		threading.GoSafeCtx(r.Context(), func() {
			defer close(client)
			err := l.Ask(&req, client)
			if err != nil {
				logc.Errorw(r.Context(), "AskHandler", logc.Field("error", err))
				return
			}
		})

		for {
			select {
			case data, ok := <-client:
				if !ok {
					return
				}
				output, err := json.Marshal(data)
				if err != nil {
					logc.Errorw(r.Context(), "AskHandler", logc.Field("error", err))
					continue
				}

				if _, err := fmt.Fprintf(w, "data: %s\n\n", string(output)); err != nil {
					logc.Errorw(r.Context(), "AskHandler", logc.Field("error", err))
					return
				}
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
			case <-r.Context().Done():
				return
			}
		}
	}
}
