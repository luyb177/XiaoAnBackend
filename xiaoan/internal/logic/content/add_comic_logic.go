package content

import (
	"context"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddComicLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAddComicLogic 添加漫画
func NewAddComicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddComicLogic {
	return &AddComicLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddComicLogic) AddComic(req *types.AddComicRequest) (resp *types.Response, err error) {
	res, err := l.svcCtx.ContentRpc.AddComic(l.ctx, &content.AddComicRequest{
		Name:        req.Name,
		Tag:         req.Tags,
		Description: req.Description,
		Cover:       req.Cover,
		Author:      req.Author,
		PublishedAt: req.PublishedAt,
	})
	if err != nil {
		return &types.Response{
			Code:    400,
			Message: err.Error(),
		}, nil
	}

	var data *content.AddComicResponse
	if res.Data != nil {
		data = &content.AddComicResponse{}
		_ = res.Data.UnmarshalTo(data)
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
