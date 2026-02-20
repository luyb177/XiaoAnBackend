package content

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
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

	rpcResp, err := l.svcCtx.ContentRPC.AddPodcast(l.ctx, &content.AddPodcastRequest{
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
		l.Errorf("rpc AddPodcast err: %s", err.Error())
		return &types.Response{
			Code:    400,
			Message: "添加播客失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	// RPC 返回数据（proto 层）
	var rpcData = &content.AddPodcastResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("unmarshal AddPodcastResponse failed: %v", err)
		}
	}

	// HTTP 返回数据（对前端稳定）
	httpData := &types.AddPodcastResponse{
		PodcastID:      rpcData.Id,
		RelationStatus: rpcData.RelationStatus,
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
