package auth

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	auth "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

type ChangeClassLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewChangeClassLogic 切换班级
func NewChangeClassLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeClassLogic {
	return &ChangeClassLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChangeClassLogic) ChangeClass(req *types.ChangeClassRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.AuthRPC.ChangeClass(l.ctx, &auth.ChangeClassRequest{InviteCode: req.InviteCode})
	if err != nil {
		l.Errorf("rpc ChangeClass error: %v", err)
		return logic.BadResponse("切换班级失败"), nil
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    &types.EmptyResponse{},
	}, nil
}
