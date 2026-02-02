package content

import (
	"context"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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
	res, err := l.svcCtx.ContentRpc.AddComment(l.ctx, &content.AddCommentRequest{
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
		return &types.Response{
			Code:    400,
			Message: err.Error(),
		}, nil
	}

	var data *content.AddCommentResponse
	if res.Data != nil {
		data = &content.AddCommentResponse{}
		_ = res.Data.UnmarshalTo(data)
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
