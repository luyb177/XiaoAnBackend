package content

import (
	"context"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSubCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetSubCommentLogic 获取子评论
func NewGetSubCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSubCommentLogic {
	return &GetSubCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSubCommentLogic) GetSubComment(req *types.GetSubCommentRequest) (resp *types.Response, err error) {
	res, err := l.svcCtx.ContentRpc.GetSubComment(l.ctx, &content.GetSubCommentRequest{
		ContentType:     req.ContentType,
		ContentId:       req.ContentId,
		ParentCommentId: req.ParentCommentId,
		Page:            req.Page,
		PageSize:        req.PageSize,
	})
	if err != nil {
		return &types.Response{
			Code:    400,
			Message: err.Error(),
		}, nil
	}

	var data *content.GetSubCommentResponse

	if res.Data != nil {
		data = &content.GetSubCommentResponse{}
		_ = res.Data.UnmarshalTo(data)
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
