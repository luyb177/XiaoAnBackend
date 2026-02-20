package content

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
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

func (l *DeleteComicChapterLogic) DeleteComicChapter(
	req *types.DeleteComicChapterRequest,
) (resp *types.Response, err error) {

	rpcResp, err := l.svcCtx.ContentRPC.DeleteComicChapter(
		l.ctx,
		&content.DeleteComicChapterRequest{
			Id: req.ComicChapterId, ComicId: req.ComicID,
		})

	if err != nil {
		l.Errorf("rpc DeleteComicChapter err: %s", err.Error())
		return &types.Response{
			Code:    400,
			Message: "删除漫画章节失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    &types.EmptyResponse{},
	}, nil
}
