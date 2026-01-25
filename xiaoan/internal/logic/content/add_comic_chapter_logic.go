package content

import (
	"context"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddComicChapterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAddComicChapterLogic 添加漫画章节
func NewAddComicChapterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddComicChapterLogic {
	return &AddComicChapterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddComicChapterLogic) AddComicChapter(req *types.AddComicChapterRequest) (resp *types.Response, err error) {
	res, err := l.svcCtx.ContentRpc.AddComicChapter(l.ctx, &content.AddComicChapterRequest{
		ComicId:     req.ComicId,
		ChapterNo:   req.ChapterNo,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		PublishedAt: req.PublishedAt,
		PageUrls:    req.PageUrls,
	})
	if err != nil {
		return &types.Response{
			Code:    400,
			Message: err.Error(),
		}, nil
	}

	var data *content.AddComicChapterResponse
	if res.Data != nil {
		data = &content.AddComicChapterResponse{}
		_ = res.Data.UnmarshalTo(data)
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
