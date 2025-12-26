package auth

import (
	"context"
	auth "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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

func (l *GenerateInviteCodeLogic) GenerateInviteCode(req *types.GenerateInviteCodeRequest) (resp *types.Response, err error) {
	// 验证 req
	if req.Creator_name == "" {
		return &types.Response{
			Code:    400,
			Message: "请填写创建者名称",
		}, nil
	}

	if req.Department == "" {
		return &types.Response{
			Code:    400,
			Message: "请填写部门名称",
		}, nil
	}

	if req.Target_role == "" {
		return &types.Response{
			Code:    400,
			Message: "请填写目标角色",
		}, nil
	}

	// todo classId

	res, _ := l.svcCtx.AuthRpc.GenerateInviteCode(l.ctx, &auth.GenerateInviteCodeRequest{
		CreatorName: req.Creator_name,
		Department:  req.Department,
		MaxUses:     req.MaxUses,
		Remark:      req.Remark,
		ExpiresAt:   req.Expires_at,
		TargetRole:  req.Target_role,
		ClassId:     req.ClassId,
	})

	var data *auth.GenerateInviteCodeResponse
	if res.Data != nil {
		data = &auth.GenerateInviteCodeResponse{}
		_ = anypb.UnmarshalTo(res.Data, data, proto.UnmarshalOptions{})
	}
	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
