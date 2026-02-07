package content

import (
	"context"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddArticleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAddArticleLogic 添加文章
func NewAddArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddArticleLogic {
	return &AddArticleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddArticleLogic) AddArticle(req *types.AddArticleRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.ContentRpc.AddArticle(l.ctx, &content.AddArticleRequest{
		Name:        req.Name,
		Description: req.Description,
		Content:     req.Content,
		Cover:       req.Cover,
		Url:         req.Url,
		PublishedAt: req.PublishedAt,
		Tags:        req.Tags,
		Author:      req.Author,
	})

	if err != nil {
		l.Errorf("rpc AddArticle err: %s", err.Error())
		return &types.Response{
			Code:    400,
			Message: "添加文章失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	// RPC 返回数据（proto 层）
	rpcData := &content.AddArticleResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("unmarshal AddArticleResponse failed: %v", err)
		}
	}

	// HTTP 返回数据（对前端稳定）
	httpData := &types.AddArticleResponse{
		ArticleId:      rpcData.Id,
		RelationStatus: rpcData.RelationStatus,
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
