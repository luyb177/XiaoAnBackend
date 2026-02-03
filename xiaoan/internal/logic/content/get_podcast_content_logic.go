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
	rpcResp, err := l.svcCtx.ContentRpc.GetPodcast(l.ctx, &content.GetPodcastRequest{Id: req.PodcastId})
	if err != nil {
		l.Errorf("rpc GetPodcast err: %v", err)
		return &types.Response{
			Code:    400,
			Message: "获取播客内容失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	// RPC 数据
	rpcData := &content.GetPodcastResponse{}
	if rpcResp.Data != nil {
		if err := rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("unmarshal GetPodcastResponse failed: %v", err)
		}
	}

	// Podcast 兜底（防 panic）
	rpcPodcast := rpcData.Podcast
	if rpcPodcast == nil {
		rpcPodcast = &content.Podcast{}
	}

	// Highlights 兜底（保证 slice 非 nil）
	rpcHighlights := rpcPodcast.Highlights
	if rpcHighlights == nil {
		rpcHighlights = []*content.PodcastHighlight{}
	}

	// HTTP highlights
	httpHighlights := make([]types.PodcastHighlightItem, len(rpcHighlights))
	for i, h := range rpcHighlights {
		httpHighlights[i] = types.PodcastHighlightItem{
			Second:    h.Second,
			Highlight: h.Highlight,
		}
	}

	// HTTP 数据（稳定结构）
	httpData := &types.GetPodcastResponse{
		Podcast: types.Podcast{
			PodcastID:      rpcPodcast.Id,
			Name:           rpcPodcast.Name,
			Tags:           rpcPodcast.Tag,
			Url:            rpcPodcast.Url,
			Description:    rpcPodcast.Description,
			Cover:          rpcPodcast.Cover,
			Author:         rpcPodcast.Author,
			Channel:        rpcPodcast.Channel,
			Status:         rpcPodcast.Status,
			PublishedAt:    rpcPodcast.PublishedAt,
			CreatedAt:      rpcPodcast.CreatedAt,
			UpdatedAt:      rpcPodcast.UpdatedAt,
			LikeCount:      rpcPodcast.LikeCount,
			ViewCount:      rpcPodcast.ViewCount,
			CollectCount:   rpcPodcast.CollectCount,
			CommentCount:   rpcPodcast.CommentCount,
			Highlights:     httpHighlights,
			LastModifiedBy: rpcPodcast.LastModifiedBy,
			RelationStatus: rpcPodcast.RelationStatus,
			IsLiked:        rpcPodcast.IsLiked,
			IsCollected:    rpcPodcast.IsCollected,
		},
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
