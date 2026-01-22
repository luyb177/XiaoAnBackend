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
	res, _ := l.svcCtx.ContentRpc.AddVideo(l.ctx, &content.AddVideoRequest{
		Name:        req.Name,
		Tag:         req.Tags,
		Url:         req.Url,
		Description: req.Description,
		Cover:       req.Cover,
		Author:      req.Author,
		PublishedAt: req.PublishedAt,
	})

	var data *content.AddVideoResponse
	if res.Data != nil {
		data = &content.AddVideoResponse{}
		_ = anypb.UnmarshalTo(res.Data, data, proto.UnmarshalOptions{})
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
