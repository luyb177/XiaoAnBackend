package content

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
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

func (l *ModifyComicChapterLogic) ModifyComicChapter(
	req *types.ModifyComicChapterRequest,
) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.ContentRPC.ModifyComicChapter(
		l.ctx,
		&content.ModifyComicChapterRequest{
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
		l.Errorf("rpc ModifyComicChapter err: %v", err)
		return &types.Response{
			Code:    400,
			Message: "修改漫画章节失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	var rpcData = &content.ModifyComicChapterResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("unmarshal ModifyComicChapterResponse failed: %v", err)
		}
	}

	httpData := &types.ModifyComicChapterResponse{
		ComicChapterId: rpcData.Id,
		RelationStatus: rpcData.RelationStatus,
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
