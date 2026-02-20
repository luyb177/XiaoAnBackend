package auth

import (
	"context"

	v1 "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type InvalidateInviteCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 失效邀请码
func NewInvalidateInviteCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InvalidateInviteCodeLogic {
	return &InvalidateInviteCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InvalidateInviteCodeLogic) InvalidateInviteCode(req *types.InvalidateInviteCodeRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.AuthRPC.InvalidateInviteCode(l.ctx, &v1.InvalidateInviteCodeRequest{Code: req.Code})

	if err != nil {
		l.Errorf("rpc InvalidateInviteCode error: %v", err)
		return logic.BadResponse("失效邀请码失败"), nil
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    &types.EmptyResponse{},
	}, nil
}
