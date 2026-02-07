package content

import (
	"context"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyComicLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewModifyComicLogic 修改漫画
func NewModifyComicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyComicLogic {
	return &ModifyComicLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ModifyComicLogic) ModifyComic(req *types.ModifyComicRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.ContentRpc.ModifyComic(l.ctx, &content.ModifyComicRequest{
		Id:          req.ComicId,
		Name:        req.Name,
		Tag:         req.Tags,
		Description: req.Description,
		Cover:       req.Cover,
		Author:      req.Author,
		PublishedAt: req.PublishedAt,
	})

	if err != nil {
		l.Errorf("rpc ModifyComic err: %v", err)
		return &types.Response{
			Code:    400,
			Message: "修改漫画失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	var rpcData = &content.ModifyComicResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("unmarshal ModifyComicResponse failed: %v", err)
		}
	}

	httpData := &types.ModifyComicResponse{
		ComicId:        rpcData.Id,
		RelationStatus: rpcData.RelationStatus,
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
