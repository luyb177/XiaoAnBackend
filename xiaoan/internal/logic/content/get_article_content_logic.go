package content

import (
	"context"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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

func (l *GetArticleContentLogic) GetArticleContent(req *types.GetArticleContentRequest) (resp *types.Response, err error) {
	if req.ArticleId <= 0 {
		l.Logger.Errorf("文章ID错误")

		return &types.Response{
			Code:    400,
			Message: "文章ID错误",
		}, nil
	}

	res, _ := l.svcCtx.ContentRpc.GetArticle(l.ctx, &content.GetArticleRequest{
		Id: req.ArticleId,
	})

	var data *content.GetArticleResponse
	if res.Data != nil {
		data = &content.GetArticleResponse{}
		_ = anypb.UnmarshalTo(res.Data, data, proto.UnmarshalOptions{})
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
