package content

import (
	"context"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

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
	// 构造
	images := make([]*content.AddArticleImage, 0, len(req.Images))
	for _, image := range req.Images {
		images = append(images, &content.AddArticleImage{
			Url:  image.Url,
			Sort: image.Sort,
			Tp:   image.Tp,
		})
	}

	res, _ := l.svcCtx.ContentRpc.AddArticle(l.ctx, &content.AddArticleRequest{
		Name:        req.Name,
		Description: req.Description,
		Content:     req.Content,
		Cover:       req.Cover,
		Url:         req.Url,
		PublishedAt: req.PublishedAt,
		Tags:        req.Tags,
		Images:      images,
		Author:      req.Author,
	})

	var data *content.AddArticleResponse
	if res.Data != nil {
		data = &content.AddArticleResponse{}
		_ = anypb.UnmarshalTo(res.Data, data, proto.UnmarshalOptions{})
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
