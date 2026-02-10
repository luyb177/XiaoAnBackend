package content

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
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
	rpcResp, err := l.svcCtx.ContentRPC.AddComicChapter(l.ctx, &content.AddComicChapterRequest{
		ComicId:     req.ComicId,
		ChapterNo:   req.ChapterNo,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		PublishedAt: req.PublishedAt,
		PageUrls:    req.PageUrls,
	})
	if err != nil {
		l.Errorf("rpc AddComicChapter error: %v", err)
		return &types.Response{
			Code:    400,
			Message: "添加漫画章节失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	var rpcData = &content.AddComicChapterResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("unmarshal AddComicChapterResponse failed: %v", err)
		}
	}

	// HTTP 返回数据（对前端稳定）
	httpData := &types.AddComicChapterResponse{
		ComicChapterId: rpcData.Id,
		RelationStatus: rpcData.RelationStatus,
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
