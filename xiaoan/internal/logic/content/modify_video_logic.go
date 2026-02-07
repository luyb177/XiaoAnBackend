package content

import (
	"context"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewModifyVideoLogic 修改视频
func NewModifyVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyVideoLogic {
	return &ModifyVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ModifyVideoLogic) ModifyVideo(req *types.ModifyVideoRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.ContentRpc.ModifyVideo(l.ctx, &content.ModifyVideoRequest{
		Id:          req.VideoId,
		Name:        req.Name,
		Tag:         req.Tags,
		Url:         req.Url,
		Description: req.Description,
		Cover:       req.Cover,
		Author:      req.Author,
		PublishedAt: req.PublishedAt,
	})
	if err != nil {
		l.Errorf("rpc ModifyVideo err: %s", err.Error())
		return &types.Response{
			Code:    400,
			Message: "修改视频失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	// RPC 返回数据（proto 层）
	var rpcData = &content.ModifyVideoResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("unmarshal ModifyVideoResponse failed: %v", err)
		}
	}

	// HTTP 返回数据（对前端稳定）
	httpData := &types.ModifyVideoResponse{
		VideoId:        rpcData.Id,
		RelationStatus: rpcData.RelationStatus,
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
