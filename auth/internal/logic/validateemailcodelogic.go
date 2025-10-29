package logic

import (
	"context"

	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pb/auth"

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

func (l *ValidateEmailCodeLogic) ValidateEmailCode(in *auth.ValidateEmailRequest) (*auth.Response, error) {
	// todo: add your logic here and delete this line

	return &auth.Response{}, nil
}
