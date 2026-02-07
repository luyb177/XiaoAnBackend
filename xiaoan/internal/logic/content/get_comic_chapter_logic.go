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
	rpcResp, err := l.svcCtx.ContentRpc.GetComicChapter(l.ctx, &content.GetComicChapterRequest{
		ComicId:  req.ComicId,
		Page:     req.Page,
		PageSize: req.PageSize,
	})

	if err != nil {
		l.Errorf("rpc GetComicChapter err: %s", err.Error())
		return &types.Response{
			Code:    400,
			Message: "获取漫画章节失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	// RPC 返回数据（proto 层）
	var rpcData = &content.GetComicChapterResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("unmarshal GetComicChapterResponse failed: %v", err)
		}
	}

	rpcChapters := rpcData.Chapters
	if rpcChapters == nil {
		rpcChapters = []*content.ComicChapter{}
	}

	// HTTP 返回数据（对前端稳定）
	httpChapters := make([]types.ComicChapter, len(rpcChapters))
	for i, chapter := range rpcChapters {
		if chapter == nil {
			chapter = &content.ComicChapter{}
		}

		httpChapters[i] = types.ComicChapter{
			ComicChapterID: chapter.Id,
			ComicID:        chapter.ComicId,
			ChapterNo:      chapter.ChapterNo,
			Title:          chapter.Title,
			Description:    chapter.Description,
			Status:         chapter.Status,
			PageCount:      chapter.PageCount,
			PublishedAt:    chapter.PublishedAt,
			CreatedAt:      chapter.CreatedAt,
			UpdatedAt:      chapter.UpdatedAt,
		}
	}

	httpData := &types.GetComicChapterResponse{
		Chapters: httpChapters,
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
