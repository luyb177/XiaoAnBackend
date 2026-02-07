package auth

import (
	"context"

	auth "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetInviteCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetInviteCodeLogic 获取邀请码
func NewGetInviteCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInviteCodeLogic {
	return &GetInviteCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetInviteCodeLogic) GetInviteCode(req *types.GetInviteCodeRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.AuthRpc.GetInviteCode(l.ctx, &auth.GetInviteCodeRequest{
		PageSize: req.PageSize,
		Cursor:   req.Cursor,
	})

	if err != nil {
		l.Errorf("rpc GetInviteCode err: %v", err)
		return &types.Response{
			Code:    400,
			Message: "获取邀请码失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	var rpcData = &auth.GetInviteCodeResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("rpc GetInviteCode UnmarshalTo err: %v", err)
		}
	}
	rpcInviteCodes := rpcData.Codes
	if rpcInviteCodes == nil {
		rpcInviteCodes = []*auth.InviteCode{}
	}

	httpInviteCodes := make([]types.InviteCode, len(rpcInviteCodes))
	for i, rpcInviteCode := range rpcInviteCodes {
		httpInviteCodes[i] = types.InviteCode{
			Code:        rpcInviteCode.Code,
			CreatorID:   rpcInviteCode.CreatorId,
			CreatorName: rpcInviteCode.CreatorName,
			Department:  rpcInviteCode.Department,
			MaxUses:     rpcInviteCode.MaxUses,
			UsedCount:   rpcInviteCode.UsedCount,
			Remark:      rpcInviteCode.Remark,
			ExpiresAt:   rpcInviteCode.ExpiresAt,
			TargetRole:  rpcInviteCode.TargetRole,
			ClassId:     rpcInviteCode.ClassId,
			CreatedAt:   rpcInviteCode.CreatedAt,
			UpdatedAt:   rpcInviteCode.UpdatedAt,
		}
	}

	httpData := types.GetInviteCodeResponse{
		InviteCodes: httpInviteCodes,
		HasMore:     rpcData.HasMore,
		NextCursor:  rpcData.NextCursor,
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
