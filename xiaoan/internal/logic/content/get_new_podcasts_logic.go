package content

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

type GetNewPodcastsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetNewPodcastsLogic 获取最新播客
func NewGetNewPodcastsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNewPodcastsLogic {
	return &GetNewPodcastsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNewPodcastsLogic) GetNewPodcasts(req *types.GetNewPodcastsRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.ContentRPC.GetNewPodcasts(l.ctx, &content.GetNewPodcastsRequest{
		PageSize: req.PageSize,
		Cursor:   req.Cursor,
	})

	if err != nil {
		l.Errorf("rpc GetNewPodcasts err: %v", err)
		return logic.BadResponse("获取最新播客失败"), nil
	}

	var rpcData = &content.GetNewPodcastsResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("rpc GetNewPodcasts UnmarshalTo err: %v", err)
		}
	}
	rpcPodcasts := rpcData.Podcasts
	if rpcPodcasts == nil {
		rpcPodcasts = []*content.Podcast{}
	}

	httpPodcasts := make([]types.PodcastInfo, len(rpcPodcasts))
	for i, rpcPodcast := range rpcPodcasts {
		if rpcPodcast == nil {
			rpcPodcast = &content.Podcast{}
		}
		httpPodcasts[i] = types.PodcastInfo{
			PodcastID:      rpcPodcast.Id,
			Name:           rpcPodcast.Name,
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
			LastModifiedBy: rpcPodcast.LastModifiedBy,
			RelationStatus: rpcPodcast.RelationStatus,
		}
	}

	httpData := &types.GetNewPodcastsResponse{
		Podcasts:   httpPodcasts,
		HasMore:    rpcData.HasMore,
		NextCursor: rpcData.NextCursor,
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
