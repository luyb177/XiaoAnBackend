package content

import (
	"context"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyArticleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 修改文章
func NewModifyArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyArticleLogic {
	return &ModifyArticleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ModifyArticleLogic) ModifyArticle(req *types.ModifyArticleRequest) (resp *types.Response, err error) {
	res, err := l.svcCtx.ContentRpc.ModifyArticle(l.ctx, &content.ModifyArticleRequest{
		Id:          req.ArticleId,
		Name:        req.Name,
		Tag:         req.Tags,
		Url:         req.Url,
		Description: req.Description,
		Cover:       req.Cover,
		Content:     req.Content,
		Author:      req.Author,
		PublishedAt: req.PublishedAt,
	})

	if err != nil {
		return &types.Response{
			Code:    400,
			Message: err.Error(),
		}, nil
	}

	var data *content.ModifyArticleResponse

	if res.Data != nil {
		data = &content.ModifyArticleResponse{}
		_ = res.Data.UnmarshalTo(data)
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
