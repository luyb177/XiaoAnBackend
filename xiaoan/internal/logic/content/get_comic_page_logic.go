package content

import (
	"context"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetComicPageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetComicPageLogic 获取漫画页面
func NewGetComicPageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetComicPageLogic {
	return &GetComicPageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetComicPageLogic) GetComicPage(req *types.GetComicPageRequest) (resp *types.Response, err error) {
	res, err := l.svcCtx.ContentRpc.GetComicPage(l.ctx, &content.GetComicPageRequest{
		ComicChapterId: req.ComicChapterId,
		Page:           req.Page,
		PageSize:       req.PageSize,
	})

	if err != nil {
		return &types.Response{
			Code:    400,
			Message: err.Error(),
		}, nil
	}

	var data *content.GetComicPageResponse
	if res.Data != nil {
		data = &content.GetComicPageResponse{}
		_ = res.Data.UnmarshalTo(data)
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
