package logic

import (
	"context"

	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pb/auth"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateInviteCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenerateInviteCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateInviteCodeLogic {
	return &GenerateInviteCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 邀请码
func (l *GenerateInviteCodeLogic) GenerateInviteCode(in *auth.GenerateInviteCodeRequest) (*auth.Response, error) {
	// todo: add your logic here and delete this line

	return &auth.Response{}, nil
}
