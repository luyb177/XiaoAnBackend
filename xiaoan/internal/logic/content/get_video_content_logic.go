package content

import (
	"context"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetVideoContentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetVideoContentLogic 获取视频详细内容
func NewGetVideoContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVideoContentLogic {
	return &GetVideoContentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetVideoContentLogic) GetVideoContent(req *types.GetVideoContentRequest) (resp *types.Response, err error) {
	res, err := l.svcCtx.ContentRpc.GetVideo(l.ctx, &content.GetVideoRequest{
		Id: req.VideoId,
	})

	if err != nil {
		return &types.Response{
			Code:    400,
			Message: err.Error(),
		}, nil
	}

	var data *content.GetVideoResponse
	if res.Data != nil {
		data = &content.GetVideoResponse{}
		_ = res.Data.UnmarshalTo(data)
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
