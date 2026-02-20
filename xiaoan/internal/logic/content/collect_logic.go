package content

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

type CollectLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 收藏
func NewCollectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CollectLogic {
	return &CollectLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CollectLogic) Collect(req *types.CollectRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.ContentRPC.Collect(l.ctx, &content.CollectRequest{
		ContentType: req.ContentType,
		ContentId:   req.ContentId,
	})

	if err != nil {
		l.Errorf("rpc Collect err: %s", err.Error())
		return &types.Response{
			Code:    400,
			Message: "收藏失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    &types.EmptyResponse{},
	}, nil
}
