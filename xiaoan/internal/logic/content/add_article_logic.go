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
	res, err := l.svcCtx.ContentRpc.AddArticle(l.ctx, &content.AddArticleRequest{
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
		return &types.Response{
			Code:    400,
			Message: err.Error(),
		}, nil
	}

	var data *content.AddArticleResponse
	if res.Data != nil {
		data = &content.AddArticleResponse{}
		_ = res.Data.UnmarshalTo(data)
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
