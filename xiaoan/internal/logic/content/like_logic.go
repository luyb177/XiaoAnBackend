package content

import (
	"context"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LikeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewLikeLogic 点赞
func NewLikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeLogic {
	return &LikeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LikeLogic) Like(req *types.LikeRequest) (resp *types.Response, err error) {
	res, err := l.svcCtx.ContentRpc.Like(l.ctx, &content.LikeRequest{
		ContentType: req.ContentType,
		ContentId:   req.ContentId,
	})

	if err != nil {
		return &types.Response{
			Code:    400,
			Message: "点赞失败",
		}, nil
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
	}, nil
}
