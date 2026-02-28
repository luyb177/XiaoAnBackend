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

type GetOrCreateSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetOrCreateSessionLogic 获取或创建问答会话
func NewGetOrCreateSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrCreateSessionLogic {
	return &GetOrCreateSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrCreateSessionLogic) GetOrCreateSession(req *types.GetOrCreateSessionRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.QARPC.GetOrCreateSession(l.ctx, &qa.GetOrCreateSessionRequest{})
	if err != nil {
		l.Errorf("call GetOrCreateSession error: %v", err)
		return logic.InternalErrorResponse("获取会话失败"), nil
	}

	var rpcData = &qa.GetOrCreateSessionResponse{}
	if rpcResp.Data != nil {
		if err := rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("unmarshal GetOrCreateSessionResponse error: %v", err)
		}
	}
	rpcChatSession := rpcData.Session
	if rpcChatSession == nil {
		rpcChatSession = &qa.ChatSession{}
	}

	httpChatSession := types.ChatSession{
		SessionID:      rpcChatSession.Id,
		Title:          rpcChatSession.Title,
		UserID:         rpcChatSession.UserId,
		MessageCount:   rpcChatSession.MessageCount,
		HasMessage:     rpcChatSession.HasMessage,
		SessionStatus:  rpcChatSession.SessionStatus,
		IsPinned:       rpcChatSession.IsPinned,
		PinnedAt:       rpcChatSession.PinnedAt,
		LastMessageAt:  rpcChatSession.LastMessageAt,
		RelationStatus: rpcChatSession.RelationStatus,
		CreatedAt:      rpcChatSession.CreatedAt,
		UpdatedAt:      rpcChatSession.UpdatedAt,
	}

	httpData := types.GetOrCreateSessionResponse{Session: httpChatSession}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
