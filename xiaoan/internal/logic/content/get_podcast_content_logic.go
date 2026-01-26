package content

import (
	"context"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPodcastContentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetPodcastContentLogic 获取播客详细内容
func NewGetPodcastContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPodcastContentLogic {
	return &GetPodcastContentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPodcastContentLogic) GetPodcastContent(req *types.GetPodcastContentRequest) (resp *types.Response, err error) {
	res, err := l.svcCtx.ContentRpc.GetPodcast(l.ctx, &content.GetPodcastRequest{Id: req.PodcastId})

	if err != nil {
		return &types.Response{
			Code:    400,
			Message: err.Error(),
		}, nil
	}

	var data *content.GetPodcastResponse
	if res.Data != nil {
		data = &content.GetPodcastResponse{}
		_ = res.Data.UnmarshalTo(data)
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
