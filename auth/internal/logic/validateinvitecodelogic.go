package logic

import (
	"context"

	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pb/auth"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateInviteCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateInviteCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateInviteCodeLogic {
	return &ValidateInviteCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ValidateInviteCodeLogic) ValidateInviteCode(in *auth.ValidateInviteCodeRequest) (*auth.Response, error) {
	// todo: add your logic here and delete this line

	return &auth.Response{}, nil
}
