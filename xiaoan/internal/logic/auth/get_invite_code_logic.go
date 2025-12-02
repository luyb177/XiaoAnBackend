package auth

import (
	"context"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetInviteCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取邀请码
func NewGetInviteCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInviteCodeLogic {
	return &GetInviteCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetInviteCodeLogic) GetInviteCode(req *types.GetInviteCodeRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}
