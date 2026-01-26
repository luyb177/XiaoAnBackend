package content

import (
	"context"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyComicLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewModifyComicLogic 修改漫画
func NewModifyComicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyComicLogic {
	return &ModifyComicLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ModifyComicLogic) ModifyComic(req *types.ModifyComicRequest) (resp *types.Response, err error) {
	res, err := l.svcCtx.ContentRpc.ModifyComic(l.ctx, &content.ModifyComicRequest{
		Id:          req.ComicId,
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

	var data *content.ModifyComicResponse
	if res.Data != nil {
		data = &content.ModifyComicResponse{}
		_ = res.Data.UnmarshalTo(data)
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
