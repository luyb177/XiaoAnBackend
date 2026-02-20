package content

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

type GetComicLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetComicLogic 获取漫画
func NewGetComicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetComicLogic {
	return &GetComicLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetComicLogic) GetComic(req *types.GetComicRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.ContentRPC.GetComic(l.ctx, &content.GetComicRequest{Id: req.ComicID})
	if err != nil {
		l.Errorf("rpc GetComic err: %s", err.Error())
		return &types.Response{
			Code:    400,
			Message: "获取漫画失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	// RPC 返回数据（proto 层）
	var rpcData = &content.GetComicResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("unmarshal GetComicResponse failed: %v", err)
		}
	}
	rpcComic := rpcData.Comic
	if rpcComic == nil {
		rpcComic = &content.Comic{}
	}

	// HTTP 返回数据（对前端稳定）
	httpData := &types.GetComicResponse{
		Comic: types.Comic{
			ComicID:        rpcComic.Id,
			Name:           rpcComic.Name,
			Tags:           rpcComic.Tag,
			Description:    rpcComic.Description,
			Cover:          rpcComic.Cover,
			Author:         rpcComic.Author,
			PublishedAt:    rpcComic.PublishedAt,
			CreatedAt:      rpcComic.CreatedAt,
			UpdatedAt:      rpcComic.UpdatedAt,
			LikeCount:      rpcComic.LikeCount,
			ViewCount:      rpcComic.ViewCount,
			CollectCount:   rpcComic.CollectCount,
			CommentCount:   rpcComic.CommentCount,
			ChapterCount:   rpcComic.ChapterCount,
			LastModifiedBy: rpcComic.LastModifiedBy,
			RelationStatus: rpcComic.RelationStatus,
			IsLiked:        rpcComic.IsLiked,
			IsCollected:    rpcComic.IsCollected,
		}}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
