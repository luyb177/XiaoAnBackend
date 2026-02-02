package content

import (
	"context"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnlikeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUnlikeLogic 取消点赞
func NewUnlikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnlikeLogic {
	return &UnlikeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnlikeLogic) Unlike(req *types.UnlikeRequest) (resp *types.Response, err error) {
	res, err := l.svcCtx.ContentRpc.Unlike(l.ctx, &content.UnlikeRequest{
		ContentType: req.ContentType,
		ContentId:   req.ContentId,
	})

	if err != nil {
		return &types.Response{
			Code:    400,
			Message: "取消点赞失败",
		}, nil
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
	}, nil
}
