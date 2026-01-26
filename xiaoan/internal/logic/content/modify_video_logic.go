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
	res, err := l.svcCtx.ContentRpc.ModifyVideo(l.ctx, &content.ModifyVideoRequest{
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
		return &types.Response{
			Code:    400,
			Message: err.Error(),
		}, nil
	}

	var data *content.ModifyVideoResponse
	if res.Data != nil {
		data = &content.ModifyVideoResponse{}
		_ = res.Data.UnmarshalTo(data)
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
