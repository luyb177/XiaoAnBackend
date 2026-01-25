package content

import (
	"context"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddPodcastLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAddPodcastLogic 添加播客
func NewAddPodcastLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddPodcastLogic {
	return &AddPodcastLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddPodcastLogic) AddPodcast(req *types.AddPodcastRequest) (resp *types.Response, err error) {
	highlights := make([]*content.PodcastHighlight, len(req.Highlights))
	for i, highlight := range req.Highlights {
		highlights[i] = &content.PodcastHighlight{
			Second:    highlight.Second,
			Highlight: highlight.Highlight,
		}
	}

	res, err := l.svcCtx.ContentRpc.AddPodcast(l.ctx, &content.AddPodcastRequest{
		Name:        req.Name,
		Url:         req.Url,
		Description: req.Description,
		Cover:       req.Cover,
		Author:      req.Author,
		PublishedAt: req.PublishedAt,
		Channel:     req.Channel,
		Status:      req.Status,
		Tags:        req.Tags,
		Highlights:  highlights,
	})

	if err != nil {
		return &types.Response{
			Code:    400,
			Message: err.Error(),
		}, nil
	}

	var data *content.AddPodcastResponse
	if res.Data != nil {
		data = &content.AddPodcastResponse{}
		_ = res.Data.UnmarshalTo(data)
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
