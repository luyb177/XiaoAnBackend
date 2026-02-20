package content

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

type DeleteCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDeleteCommentLogic 删除评论
func NewDeleteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCommentLogic {
	return &DeleteCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteCommentLogic) DeleteComment(req *types.DeleteCommentRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.ContentRPC.DeleteComment(l.ctx, &content.DeleteCommentRequest{Id: req.CommentID})
	if err != nil {
		l.Errorf("rpc DeleteComment err: %s", err.Error())
		return &types.Response{
			Code:    400,
			Message: "删除评论失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    &types.EmptyResponse{},
	}, nil
}
