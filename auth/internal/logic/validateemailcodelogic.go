package logic

import (
	"context"

	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateEmailCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateEmailCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateEmailCodeLogic {
	return &ValidateEmailCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ValidateEmailCodeLogic) ValidateEmailCode(in *v1.ValidateEmailRequest) (*v1.Response, error) {

	return &v1.Response{}, nil
}
