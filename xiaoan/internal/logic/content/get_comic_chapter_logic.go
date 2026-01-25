package content

import (
	"context"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetComicChapterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetComicChapterLogic 获取漫画章节
func NewGetComicChapterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetComicChapterLogic {
	return &GetComicChapterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetComicChapterLogic) GetComicChapter(req *types.GetComicChapterRequest) (resp *types.Response, err error) {
	res, err := l.svcCtx.ContentRpc.GetComicChapter(l.ctx, &content.GetComicChapterRequest{
		ComicId:  req.ComicId,
		Page:     req.Page,
		PageSize: req.PageSize,
	})

	if err != nil {
		return &types.Response{
			Code:    400,
			Message: err.Error(),
		}, nil
	}

	var data *content.GetComicChapterResponse
	if res.Data != nil {
		data = &content.GetComicChapterResponse{}
		_ = res.Data.UnmarshalTo(data)
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
