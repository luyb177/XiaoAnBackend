package auth

import (
	"context"

	auth "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewRegisterLogic 注册
func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.AuthRpc.Register(l.ctx, &auth.RegisterRequest{
		Email:          req.Email,
		EmailCode:      req.EmailCode,
		Password:       req.Password,
		InviteCodeUsed: req.InviteCodeUsed,
	})

	if err != nil {
		l.Errorf("rpc Register err: %v", err)
		return &types.Response{
			Code:    400,
			Message: "注册失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	var rpcData = &auth.RegisterResponse{}
	if rpcResp.Data != nil {
		if err := rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("rpc Register UnmarshalTo err: %v", err)
		}
	}
	rpcUser := rpcData.User
	if rpcUser == nil {
		rpcUser = &auth.User{}
	}

	// HTTP 返回数据（对前端稳定）
	httpUser := types.User{
		UserID:         rpcUser.Id,
		Name:           rpcUser.Name,
		Email:          rpcUser.Email,
		Avatar:         rpcUser.Avatar,
		Phone:          rpcUser.Phone,
		Department:     rpcUser.Department,
		Role:           rpcUser.Role,
		ClassID:        rpcUser.ClassId,
		Status:         rpcUser.Status,
		InviteCodeUsed: rpcUser.InviteCodeUsed,
		CreatedAt:      rpcUser.CreatedAt,
		UpdatedAt:      rpcUser.UpdatedAt,
	}

	httpData := &types.RegisterResponse{User: httpUser}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
