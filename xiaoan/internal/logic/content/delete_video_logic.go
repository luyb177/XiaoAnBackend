package content

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
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
	rpcResp, err := l.svcCtx.ContentRPC.DeleteVideo(l.ctx, &content.DeleteVideoRequest{Id: req.VideoId})
	if err != nil {
		l.Errorf("rpc DeleteVideo err: %s", err.Error())
		return &types.Response{
			Code:    400,
			Message: "删除视频失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    &types.EmptyResponse{},
	}, nil
}
