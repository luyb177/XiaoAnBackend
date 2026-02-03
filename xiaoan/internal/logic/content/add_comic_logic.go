package content

import (
	"context"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddComicLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAddComicLogic 添加漫画
func NewAddComicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddComicLogic {
	return &AddComicLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddComicLogic) AddComic(req *types.AddComicRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.ContentRpc.AddComic(l.ctx, &content.AddComicRequest{
		Name:        req.Name,
		Tag:         req.Tags,
		Description: req.Description,
		Cover:       req.Cover,
		Author:      req.Author,
		PublishedAt: req.PublishedAt,
	})
	if err != nil {
		l.Errorf("rpc AddComic error: %v", err)
		return &types.Response{
			Code:    400,
			Message: "添加漫画失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	// RPC 返回数据（proto 层）
	var rpcData = &content.AddComicResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("unmarshal AddComicResponse failed: %v", err)
		}
	}

	// HTTP 返回数据（对前端稳定）
	httpData := &types.AddComicResponse{
		ComicId:        rpcData.Id,
		RelationStatus: rpcData.RelationStatus,
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
