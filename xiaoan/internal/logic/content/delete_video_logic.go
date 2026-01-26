package content

import (
	"context"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDeleteVideoLogic 删除视频
func NewDeleteVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteVideoLogic {
	return &DeleteVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteVideoLogic) DeleteVideo(req *types.DeleteVideoRequest) (resp *types.Response, err error) {
	res, err := l.svcCtx.ContentRpc.DeleteVideo(l.ctx, &content.DeleteVideoRequest{Id: req.VideoId})
	if err != nil {
		return &types.Response{
			Code:    400,
			Message: err.Error(),
		}, nil
	}
	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
	}, nil
}
