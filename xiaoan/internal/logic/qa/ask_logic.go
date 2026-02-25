// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package qa

import (
	"context"
	"errors"

	qa "github.com/luyb177/XiaoAnBackend/qa/pb/qa/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAskLogic 流式问答接口
func NewAskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AskLogic {
	return &AskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AskLogic) Ask(req *types.AskRequest, client chan<- *types.AskStreamReply) error {
	// 先发一个初始消息，表示问答开始了
	client <- &types.AskStreamReply{
		Delta:    "",
		Finished: false,
		Code:     200,
		Message:  "start",
	}

	stream, err := l.svcCtx.QARPC.Ask(l.ctx, &qa.AskRequest{
		SessionId:       req.SessionID,
		Content:         req.Content,
		ClientMessageId: req.ClientMessageID,
	})
	if err != nil {
		l.Errorf("Ask QA rpc call err: %v", err)
		return err
	}

	for {
		res, err := stream.Recv()
		if err != nil {
			// 正常结束不要当错误打爆日志
			if errors.Is(err, context.Canceled) {
				return nil
			}
			// io.EOF 是正常结束
			if err.Error() == "EOF" {
				return nil
			}

			l.Errorf("Ask QA rpc recv err: %v", err)
			return err
		}

		client <- &types.AskStreamReply{
			Delta:    res.Delta,
			Code:     res.Code,
			Message:  res.Message,
			Finished: res.Finished,
		}

		// 如果后端明确 finished，可以提前收
		if res.Finished {
			return nil
		}
	}
}
