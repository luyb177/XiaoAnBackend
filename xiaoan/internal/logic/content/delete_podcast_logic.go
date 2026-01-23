package content

import (
	"context"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeletePodcastLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDeletePodcastLogic 删除播客
func NewDeletePodcastLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePodcastLogic {
	return &DeletePodcastLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeletePodcastLogic) DeletePodcast(req *types.DeletePodcastRequest) (resp *types.Response, err error) {
	res, _ := l.svcCtx.ContentRpc.DeletePodcast(l.ctx, &content.DeletePodcastRequest{
		Id: req.PodcastId,
	})

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
	}, nil
}
