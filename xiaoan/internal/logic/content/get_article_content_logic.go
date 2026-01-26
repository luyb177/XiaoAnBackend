package content

import (
	"context"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetArticleContentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取文章详细内容
func NewGetArticleContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetArticleContentLogic {
	return &GetArticleContentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetArticleContentLogic) GetArticleContent(req *types.GetArticleContentRequest) (resp *types.Response, err error) {
	res, err := l.svcCtx.ContentRpc.GetArticle(l.ctx, &content.GetArticleRequest{
		Id: req.ArticleId,
	})

	if err != nil {
		return &types.Response{
			Code:    400,
			Message: err.Error(),
		}, nil
	}

	var data *content.GetArticleResponse
	if res.Data != nil {
		data = &content.GetArticleResponse{}
		_ = res.Data.UnmarshalTo(data)
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
