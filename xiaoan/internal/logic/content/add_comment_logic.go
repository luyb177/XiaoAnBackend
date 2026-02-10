package content

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

type AddCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAddCommentLogic 添加评论
func NewAddCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddCommentLogic {
	return &AddCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddCommentLogic) AddComment(req *types.AddCommentRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.ContentRPC.AddComment(l.ctx, &content.AddCommentRequest{
		Type:           req.ContentType,
		TargetId:       req.ContentId,
		Nickname:       req.UserName,
		Avatar:         req.Avatar,
		ParentId:       req.ParentId,
		ReplyCommentId: req.ReplyCommentId,
		ReplyUserId:    req.ReplyUserId,
		Content:        req.CommentText,
		Status:         req.Status,
	})
	if err != nil {
		l.Errorf("rpc AddComment error: %v", err)
		return &types.Response{
			Code:    400,
			Message: "添加评论失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	var rpcData = &content.AddCommentResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("rpc AddComment unmarshal error: %v", err)
		}
	}

	httpData := &types.AddCommentResponse{
		CommentId:      rpcData.Id,
		RelationStatus: rpcData.RelationStatus,
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
