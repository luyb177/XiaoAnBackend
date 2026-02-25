// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package qa

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	qa "github.com/luyb177/XiaoAnBackend/qa/pb/qa/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

type GetMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetMessagesLogic 获取问答消息列表
func NewGetMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMessagesLogic {
	return &GetMessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMessagesLogic) GetMessages(req *types.GetMessagesRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.QARPC.GetMessageList(l.ctx, &qa.GetMessageListRequest{
		SessionId: req.SessionID,
		Cursor:    req.Cursor,
		PageSize:  req.PageSize,
		UserId:    req.UserID,
	})

	if err != nil {
		l.Errorf("GetMessageList QA rpc call err: %v", err)
		return logic.InternalErrorResponse("获取消息列表失败"), nil
	}

	var rpcData = &qa.GetMessageListResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("GetMessageList QA rpc data unmarshal err: %v", err)
		}
	}

	rpcMessages := rpcData.Messages
	if rpcMessages == nil {
		rpcMessages = []*qa.ChatMessage{}
	}

	httpMessages := make([]types.ChatMessage, len(rpcMessages))
	for i, rpcMessage := range rpcMessages {
		if rpcMessage == nil {
			rpcMessage = &qa.ChatMessage{}
		}
		httpMessages[i] = types.ChatMessage{
			MessageID:        rpcMessage.MessageId,
			SessionID:        rpcMessage.SessionId,
			UserID:           rpcMessage.UserId,
			Status:           rpcMessage.Status,
			PromptTokens:     rpcMessage.PromptTokens,
			CompletionTokens: rpcMessage.CompletionTokens,
			TotalTokens:      rpcMessage.TotalTokens,
			MessageType:      rpcMessage.MessageType,
			Content:          rpcMessage.Content,
			FinishReason:     rpcMessage.FinishReason,
			CreatedAt:        rpcMessage.CreatedAt,
			UpdatedAt:        rpcMessage.UpdatedAt,
		}
	}

	httpData := types.GetMessagesResponse{
		Messages:   httpMessages,
		HasMore:    rpcData.HasMore,
		NextCursor: rpcData.NextCursor,
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
