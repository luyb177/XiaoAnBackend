package content

import (
	"context"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyComicChapterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewModifyComicChapterLogic 修改漫画章节
func NewModifyComicChapterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyComicChapterLogic {
	return &ModifyComicChapterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ModifyComicChapterLogic) ModifyComicChapter(req *types.ModifyComicChapterRequest) (resp *types.Response, err error) {
	res, err := l.svcCtx.ContentRpc.ModifyComicChapter(l.ctx, &content.ModifyComicChapterRequest{
		Id:          req.ComicChapterId,
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

	var data *content.ModifyComicChapterResponse
	if res.Data != nil {
		data = &content.ModifyComicChapterResponse{}
		_ = res.Data.UnmarshalTo(data)
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
