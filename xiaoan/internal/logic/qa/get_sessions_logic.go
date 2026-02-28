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

type GetSessionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetSessionsLogic 获取问答会话列表
func NewGetSessionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSessionsLogic {
	return &GetSessionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSessionsLogic) GetSessions(req *types.GetSessionsRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.QARPC.GetSessionList(l.ctx, &qa.GetSessionListRequest{
		Cursor:   req.Cursor,
		PageSize: req.PageSize,
		UserId:   req.UserID,
	})

	if err != nil {
		l.Errorf("GetSessionList QA rpc call err: %v", err)
		return logic.InternalErrorResponse("获取会话列表失败"), nil
	}

	var rpcData = &qa.GetSessionListResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("GetSessionList QA rpc data unmarshal err: %v", err)
		}
	}

	rpcSessions := rpcData.Sessions
	if rpcSessions == nil {
		rpcSessions = []*qa.ChatSession{}
	}

	httpSessions := make([]types.ChatSession, len(rpcSessions))

	for i, rpcSession := range rpcSessions {
		if rpcSession == nil {
			rpcSession = &qa.ChatSession{}
		}
		httpSessions[i] = types.ChatSession{
			SessionID:      rpcSession.Id,
			Title:          rpcSession.Title,
			UserID:         rpcSession.UserId,
			MessageCount:   rpcSession.MessageCount,
			HasMessage:     rpcSession.HasMessage,
			SessionStatus:  rpcSession.SessionStatus,
			IsPinned:       rpcSession.IsPinned,
			PinnedAt:       rpcSession.PinnedAt,
			LastMessageAt:  rpcSession.LastMessageAt,
			RelationStatus: rpcSession.RelationStatus,
			CreatedAt:      rpcSession.CreatedAt,
			UpdatedAt:      rpcSession.UpdatedAt,
		}
	}

	httpData := types.GetSessionsResponse{
		Sessions:   httpSessions,
		HasMore:    rpcData.HasMore,
		NextCursor: rpcData.NextCursor,
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
