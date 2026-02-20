package content

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

type GetNewComicsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetNewComicsLogic 获取最新漫画
func NewGetNewComicsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNewComicsLogic {
	return &GetNewComicsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNewComicsLogic) GetNewComics(req *types.GetNewComicsRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.ContentRPC.GetNewComics(l.ctx, &content.GetNewComicsRequest{
		PageSize: req.PageSize,
		Cursor:   req.Cursor,
	})

	if err != nil {
		l.Errorf("rpc GetNewComics err: %v", err)
		return logic.BadResponse("获取最新漫画失败"), nil
	}

	var rpcData = &content.GetNewComicsResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("rpc GetNewComics UnmarshalTo err: %v", err)
		}
	}
	rpcComics := rpcData.Comics
	if rpcComics == nil {
		rpcComics = []*content.Comic{}
	}

	httpComics := make([]types.ComicInfo, len(rpcComics))
	for i, rpcComic := range rpcComics {
		if rpcComic == nil {
			rpcComic = &content.Comic{}
		}
		httpComics[i] = types.ComicInfo{
			ComicID:        rpcComic.Id,
			Name:           rpcComic.Name,
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
		}
	}

	httpData := types.GetNewComicsResponse{
		Comics:     httpComics,
		HasMore:    rpcData.HasMore,
		NextCursor: rpcData.NextCursor,
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
