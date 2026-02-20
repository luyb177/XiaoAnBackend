package content

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

type GetArticleContentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetArticleContentLogic 获取文章详细内容
func NewGetArticleContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetArticleContentLogic {
	return &GetArticleContentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetArticleContentLogic) GetArticleContent(
	req *types.GetArticleContentRequest,
) (resp *types.Response, err error) {

	rpcResp, err := l.svcCtx.ContentRPC.GetArticle(
		l.ctx,
		&content.GetArticleRequest{
			Id: req.ArticleId,
		})

	if err != nil {
		l.Errorf("rpc GetArticle err: %s", err.Error())
		return &types.Response{
			Code:    400,
			Message: "获取文章内容失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	// RPC 返回数据（proto 层）
	var rpcData = &content.GetArticleResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("unmarshal GetArticleResponse failed: %v", err)
		}
	}
	rpcArticle := rpcData.Article
	if rpcArticle == nil {
		rpcArticle = &content.Article{}
	}

	// HTTP 返回数据（对前端稳定）
	httpData := &types.GetArticleResponse{
		Article: types.Article{
			ArticleID:      rpcArticle.Id,
			Name:           rpcArticle.Name,
			Tags:           rpcArticle.Tag,
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
			IsLiked:        rpcArticle.IsLiked,
			IsCollected:    rpcArticle.IsCollected,
		},
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
