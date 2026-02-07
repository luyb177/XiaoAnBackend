package content

import (
	"context"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnCollectLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUnCollectLogic 取消收藏
func NewUnCollectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnCollectLogic {
	return &UnCollectLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnCollectLogic) UnCollect(req *types.UnCollectRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.ContentRpc.UnCollect(l.ctx, &content.UnCollectRequest{
		ContentType: req.ContentType,
		ContentId:   req.ContentId,
	})
	if err != nil {
		l.Errorf("rpc UnCollect err: %s", err.Error())
		return &types.Response{
			Code:    400,
			Message: "取消收藏失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    &types.EmptyResponse{},
	}, nil
}
