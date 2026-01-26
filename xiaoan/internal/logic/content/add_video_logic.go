package content

import (
	"context"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAddVideoLogic 添加视频
func NewAddVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddVideoLogic {
	return &AddVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddVideoLogic) AddVideo(req *types.AddVideoRequest) (resp *types.Response, err error) {
	res, err := l.svcCtx.ContentRpc.AddVideo(l.ctx, &content.AddVideoRequest{
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

	var data *content.AddVideoResponse
	if res.Data != nil {
		data = &content.AddVideoResponse{}
		_ = res.Data.UnmarshalTo(data)
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
