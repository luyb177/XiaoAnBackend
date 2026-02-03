package content

import (
	"context"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetVideoContentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetVideoContentLogic 获取视频详细内容
func NewGetVideoContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVideoContentLogic {
	return &GetVideoContentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetVideoContentLogic) GetVideoContent(req *types.GetVideoContentRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.ContentRpc.GetVideo(l.ctx, &content.GetVideoRequest{
		Id: req.VideoId,
	})

	if err != nil {
		l.Errorf("rpc GetVideo err: %s", err.Error())
		return &types.Response{
			Code:    400,
			Message: "获取视频内容失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	// RPC 返回数据（proto 层）
	var rpcData = &content.GetVideoResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("unmarshal GetVideoResponse failed: %v", err)
		}
	}
	rpcVideo := rpcData.Video
	if rpcVideo == nil {
		rpcVideo = &content.Video{}
	}

	// HTTP 返回数据（对前端稳定）
	httpData := &types.GetVideoResponse{Video: types.Video{
		VideoID:        rpcVideo.Id,
		Name:           rpcVideo.Name,
		Tags:           rpcVideo.Tag,
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
		IsLiked:        rpcVideo.IsLiked,
		IsCollected:    rpcVideo.IsCollected,
	}}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
