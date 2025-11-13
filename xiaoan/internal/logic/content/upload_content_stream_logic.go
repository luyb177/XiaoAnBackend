package content

import (
	"context"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadContentStreamLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUploadContentStreamLogic 上传文件（流式传输到gRPC）
func NewUploadContentStreamLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadContentStreamLogic {
	return &UploadContentStreamLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadContentStreamLogic) UploadContentStream(req *types.UploadContentRequest) (resp *types.Response, err error) {

	return
}
