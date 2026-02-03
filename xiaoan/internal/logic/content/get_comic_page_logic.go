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
	rpcResp, err := l.svcCtx.ContentRpc.GetComicPage(l.ctx, &content.GetComicPageRequest{
		ComicChapterId: req.ComicChapterId,
		Page:           req.Page,
		PageSize:       req.PageSize,
	})

	if err != nil {
		l.Errorf("rpc GetComicPage err: %s", err.Error())
		return &types.Response{
			Code:    400,
			Message: "获取漫画页面失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	// RPC 返回数据（proto 层）
	var rpcData = &content.GetComicPageResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("unmarshal GetComicPageResponse failed: %v", err)
		}
	}
	rpcPages := rpcData.Pages
	if rpcPages == nil {
		rpcPages = []*content.ComicPage{}
	}

	// HTTP 返回数据（对前端稳定）
	httpPages := make([]types.ComicChapterPage, len(rpcPages))
	for i, rpcPage := range rpcPages {
		if rpcPage == nil {
			rpcPage = &content.ComicPage{}
		}
		httpPages[i] = types.ComicChapterPage{
			ComicChapterPageID: rpcPage.Id,
			ComicChapterID:     rpcPage.ComicChapterId,
			PageNo:             rpcPage.PageNo,
			PageURL:            rpcPage.Url,
			CreatedAt:          rpcPage.CreatedAt,
			UpdatedAt:          rpcPage.UpdatedAt,
		}
	}

	httpData := &types.GetComicChapterPageResponse{Pages: httpPages}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
