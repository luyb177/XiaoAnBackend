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
	images := make([]*content.ArticleImage, len(req.Images))
	for i, v := range req.Images {
		images[i] = &content.ArticleImage{
			Url:  v.Url,
			Sort: v.Sort,
			Tp:   v.Tp,
		}
	}

	res, _ := l.svcCtx.ContentRpc.ModifyArticle(l.ctx, &content.ModifyArticleRequest{
		Id:          req.ArticleId,
		Name:        req.Name,
		Tag:         req.Tags,
		Images:      images,
		Url:         req.Url,
		Description: req.Description,
		Cover:       req.Cover,
		Content:     req.Content,
		Author:      req.Author,
		PublishedAt: req.PublishedAt,
	})

	var data *content.ModifyArticleResponse

	if res.Data != nil {
		data = &content.ModifyArticleResponse{}
		_ = anypb.UnmarshalTo(res.Data, data, proto.UnmarshalOptions{})
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
