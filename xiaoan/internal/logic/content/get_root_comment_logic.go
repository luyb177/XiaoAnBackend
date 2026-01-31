package content

import (
	"context"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRootCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取根评论
func NewGetRootCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRootCommentLogic {
	return &GetRootCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRootCommentLogic) GetRootComment(req *types.GetRootCommentRequest) (resp *types.Response, err error) {
	res, err := l.svcCtx.ContentRpc.GetRootComment(l.ctx, &content.GetRootCommentRequest{
		ContentType: req.ContentType,
		ContentId:   req.ContentId,
		Page:        req.Page,
		PageSize:    req.PageSize,
	})

	if err != nil {
		return &types.Response{
			Code:    400,
			Message: err.Error(),
		}, nil
	}

	var data *content.GetCommentResponse
	if res.Data != nil {
		data = &content.GetCommentResponse{}
		_ = res.Data.UnmarshalTo(data)
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
