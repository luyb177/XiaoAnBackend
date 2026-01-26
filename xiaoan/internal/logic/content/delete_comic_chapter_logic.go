package content

import (
	"context"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteComicChapterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDeleteComicChapterLogic  删除漫画章节
func NewDeleteComicChapterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteComicChapterLogic {
	return &DeleteComicChapterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteComicChapterLogic) DeleteComicChapter(req *types.DeleteComicChapterRequest) (resp *types.Response, err error) {
	res, err := l.svcCtx.ContentRpc.DeleteComicChapter(l.ctx, &content.DeleteComicChapterRequest{Id: req.ComicChapterId, ComicId: req.ComicId})

	if err != nil {
		return &types.Response{
			Code:    400,
			Message: err.Error(),
		}, nil
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
	}, nil
}
