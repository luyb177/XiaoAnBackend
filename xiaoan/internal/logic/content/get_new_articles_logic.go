package content

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

type GetNewArticlesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetNewArticlesLogic 获取最新文章
func NewGetNewArticlesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNewArticlesLogic {
	return &GetNewArticlesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNewArticlesLogic) GetNewArticles(req *types.GetNewArticlesRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.ContentRPC.GetNewArticles(l.ctx, &content.GetNewArticlesRequest{
		PageSize: req.PageSize,
		Cursor:   req.Cursor,
	})

	if err != nil {
		l.Errorf("rpc GetNewArticles err: %v", err)
		return logic.BadResponse("获取最新文章失败"), nil
	}

	var rpcData = &content.GetNewArticlesResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("rpc GetNewArticles UnmarshalTo err: %v", err)
		}
	}
	rpcArticles := rpcData.Articles
	if rpcArticles == nil {
		rpcArticles = []*content.Article{}
	}

	httpArticles := make([]types.ArticleInfo, len(rpcArticles))
	for i, rpcArticle := range rpcArticles {
		if rpcArticle == nil {
			rpcArticle = &content.Article{}
		}
		httpArticles[i] = types.ArticleInfo{
			ArticleID:      rpcArticle.Id,
			Name:           rpcArticle.Name,
			Url:            rpcArticle.Url,
			Description:    rpcArticle.Description,
			Cover:          rpcArticle.Cover,
			Content:        rpcArticle.Content,
			Author:         rpcArticle.Author,
			PublishedAt:    rpcArticle.PublishedAt,
			CreatedAt:      rpcArticle.CreatedAt,
			UpdatedAt:      rpcArticle.UpdatedAt,
			LikeCount:      rpcArticle.LikeCount,
			ViewCount:      rpcArticle.ViewCount,
			CollectCount:   rpcArticle.CollectCount,
			CommentCount:   rpcArticle.CommentCount,
			LastModifiedBy: rpcArticle.LastModifiedBy,
			RelationStatus: rpcArticle.RelationStatus,
		}
	}

	httpData := types.GetNewArticlesResponse{
		Articles:   httpArticles,
		HasMore:    rpcData.HasMore,
		NextCursor: rpcData.NextCursor,
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
