package content

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

type GetNewVideosLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetNewVideosLogic 获取最新视频
func NewGetNewVideosLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNewVideosLogic {
	return &GetNewVideosLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNewVideosLogic) GetNewVideos(req *types.GetNewVideosRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.ContentRPC.GetNewVideos(l.ctx, &content.GetNewVideosRequest{
		PageSize: req.PageSize,
		Cursor:   req.Cursor,
	})
	if err != nil {
		l.Errorf("rpc GetNewVideos err: %v", err)
		return logic.BadResponse("获取最新视频失败"), nil
	}

	var rpcData = &content.GetNewVideosResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("rpc GetNewVideos UnmarshalTo err: %v", err)
		}
	}
	rpcVideos := rpcData.Videos
	if rpcVideos == nil {
		rpcVideos = []*content.Video{}
	}

	httpVideos := make([]types.VideoInfo, len(rpcVideos))
	for i, rpcVideo := range rpcVideos {
		if rpcVideo == nil {
			rpcVideo = &content.Video{}
		}
		httpVideos[i] = types.VideoInfo{
			VideoID:        rpcVideo.Id,
			Name:           rpcVideo.Name,
			Url:            rpcVideo.Url,
			Description:    rpcVideo.Description,
			Cover:          rpcVideo.Cover,
			Author:         rpcVideo.Author,
			PublishedAt:    rpcVideo.PublishedAt,
			CreatedAt:      rpcVideo.CreatedAt,
			UpdatedAt:      rpcVideo.UpdatedAt,
			LikeCount:      rpcVideo.LikeCount,
			ViewCount:      rpcVideo.ViewCount,
			CollectCount:   rpcVideo.CollectCount,
			CommentCount:   rpcVideo.CommentCount,
			LastModifiedBy: rpcVideo.LastModifiedBy,
			RelationStatus: rpcVideo.RelationStatus,
		}
	}

	httpData := types.GetNewVideosResponse{
		Videos:     httpVideos,
		HasMore:    rpcData.HasMore,
		NextCursor: rpcData.NextCursor,
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
