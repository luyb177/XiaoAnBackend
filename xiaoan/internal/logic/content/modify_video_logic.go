package content

import (
	"context"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

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
	res, _ := l.svcCtx.ContentRpc.ModifyVideo(l.ctx, &content.ModifyVideoRequest{
		Id:          req.VideoId,
		Name:        req.Name,
		Tag:         req.Tags,
		Url:         req.Url,
		Description: req.Description,
		Cover:       req.Cover,
		Author:      req.Author,
		PublishedAt: req.PublishedAt,
	})

	var data *content.ModifyVideoResponse
	if res.Data != nil {
		data = &content.ModifyVideoResponse{}
		_ = anypb.UnmarshalTo(res.Data, data, proto.UnmarshalOptions{})
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
