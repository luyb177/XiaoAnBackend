package content

import (
	"context"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetComicLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetComicLogic 获取漫画
func NewGetComicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetComicLogic {
	return &GetComicLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetComicLogic) GetComic(req *types.GetComicRequest) (resp *types.Response, err error) {
	res, err := l.svcCtx.ContentRpc.GetComic(l.ctx, &content.GetComicRequest{Id: req.ComicId})
	if err != nil {
		return &types.Response{
			Code:    400,
			Message: err.Error(),
		}, nil
	}

	var data *content.GetComicResponse
	if res.Data != nil {
		data = &content.GetComicResponse{}
		_ = res.Data.UnmarshalTo(data)
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
