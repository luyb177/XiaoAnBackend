package auth

import (
	"context"

	auth "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	LoginTypeEmailCode = "email_code"
	LoginTypePassword  = "password"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewLoginLogic 登录
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.Response, err error) {
	switch req.Tp {
	case LoginTypePassword:
		return l.LoginByPassword(req)
	case LoginTypeEmailCode:
		return l.LoginByEmailCode(req)
	default:
		return logic.BadResponse("不支持的登录类型"), nil
	}
}
func (l *LoginLogic) LoginByPassword(req *types.LoginRequest) (*types.Response, error) {
	return l.loginByRpc(&auth.LoginRequest{
		Type:     auth.LoginType_PASSWORD,
		Email:    req.Email,
		Password: req.Password,
	})
}

func (l *LoginLogic) LoginByEmailCode(req *types.LoginRequest) (*types.Response, error) {
	return l.loginByRpc(&auth.LoginRequest{
		Type:      auth.LoginType_EMAIL_CODE,
		Email:     req.Email,
		EmailCode: req.EmailCode,
	})
}

func (l *LoginLogic) loginByRpc(rpcReq *auth.LoginRequest) (*types.Response, error) {
	rpcResp, err := l.svcCtx.AuthRpc.Login(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("rpc Login err: %v", err)
		return logic.BadResponse("登录失败，请稍后重试"), nil
	}

	var rpcData auth.LoginResponse
	if rpcResp.Data != nil {
		if err := rpcResp.Data.UnmarshalTo(&rpcData); err != nil {
			l.Errorf("rpc Login UnmarshalTo err: %v", err)
		}
	}

	httpUser := convertUser(rpcData.User)

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data: &types.LoginResponse{
			Token: rpcData.Token,
			User:  httpUser,
		},
	}, nil
}

func convertUser(u *auth.User) types.User {
	if u == nil {
		return types.User{}
	}

	return types.User{
		UserID:         u.Id,
		Name:           u.Name,
		Email:          u.Email,
		Avatar:         u.Avatar,
		Phone:          u.Phone,
		Department:     u.Department,
		Role:           u.Role,
		ClassID:        u.ClassId,
		Status:         u.Status,
		InviteCodeUsed: u.InviteCodeUsed,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
}
