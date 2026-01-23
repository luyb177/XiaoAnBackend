package content

import (
	"context"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type GetPodcastContentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetPodcastContentLogic 获取播客详细内容
func NewGetPodcastContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPodcastContentLogic {
	return &GetPodcastContentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPodcastContentLogic) GetPodcastContent(req *types.GetPodcastContentRequest) (resp *types.Response, err error) {
	res, _ := l.svcCtx.ContentRpc.GetPodcast(l.ctx, &content.GetPodcastRequest{Id: req.PodcastId})

	var data *content.GetPodcastResponse
	if res.Data != nil {
		data = &content.GetPodcastResponse{}
		_ = anypb.UnmarshalTo(res.Data, data, proto.UnmarshalOptions{})
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
