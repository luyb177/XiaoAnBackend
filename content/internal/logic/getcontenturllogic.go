package logic

import (
	"context"

	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetContentURLLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetContentURLLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetContentURLLogic {
	return &GetContentURLLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetContentURL 获取访问URL
func (l *GetContentURLLogic) GetContentURL(in *v1.GetContentURLRequest) (*v1.Response, error) {

	return &v1.Response{}, nil
}
