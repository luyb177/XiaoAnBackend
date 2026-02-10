package auth

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	auth "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

type GenerateInviteCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGenerateInviteCodeLogic 生成邀请码
func NewGenerateInviteCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateInviteCodeLogic {
	return &GenerateInviteCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateInviteCodeLogic) GenerateInviteCode(
	req *types.GenerateInviteCodeRequest,
) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.AuthRPC.GenerateInviteCode(
		l.ctx, &auth.GenerateInviteCodeRequest{
			Department: req.Department,
			MaxUses:    req.MaxUses,
			Remark:     req.Remark,
			ExpiresAt:  req.ExpiresAt,
			TargetRole: req.TargetRole,
			ClassId:    req.ClassId,
		})

	if err != nil {
		l.Errorf("rpc GenerateInviteCode err: %v", err)
		return logic.BadResponse("生成邀请码失败"), nil
	}

	var rpcData = &auth.GenerateInviteCodeResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("rpc GenerateInviteCode UnmarshalTo err: %v", err)
		}
	}
	rpcInviteCode := rpcData.Code
	if rpcInviteCode == nil {
		rpcInviteCode = &auth.InviteCode{}
	}

	httpInviteCode := types.InviteCode{
		Code:       rpcInviteCode.Code,
		CreatorID:  rpcInviteCode.CreatorId,
		Department: rpcInviteCode.Department,
		MaxUses:    rpcInviteCode.MaxUses,
		UsedCount:  rpcInviteCode.UsedCount,
		Remark:     rpcInviteCode.Remark,
		ExpiresAt:  rpcInviteCode.ExpiresAt,
		TargetRole: rpcInviteCode.TargetRole,
		ClassId:    rpcInviteCode.ClassId,
		CreatedAt:  rpcInviteCode.CreatedAt,
		UpdatedAt:  rpcInviteCode.UpdatedAt,
	}

	httpData := types.GenerateInviteCodeResponse{
		InviteCode: httpInviteCode,
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
