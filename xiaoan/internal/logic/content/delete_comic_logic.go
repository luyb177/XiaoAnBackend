package content

import (
	"context"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteComicLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDeleteComicLogic 删除漫画
func NewDeleteComicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteComicLogic {
	return &DeleteComicLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteComicLogic) DeleteComic(req *types.DeleteComicRequest) (resp *types.Response, err error) {
	res, err := l.svcCtx.ContentRpc.DeleteComic(l.ctx, &content.DeleteComicRequest{Id: req.ComicId})

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
