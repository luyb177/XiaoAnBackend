package auth

import (
	"context"
	"fmt"

	auth "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
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
	if req.Email == "" {
		return &types.Response{
			Code:    400,
			Message: "邮箱不能为空",
		}, nil
	}

	switch req.Tp {
	case LoginTypePassword:
		return l.LoginByPassword(req)
	case LoginTypeEmailCode:
		return l.LoginByEmailCode(req)
	default:
		return &types.Response{
			Code:    400,
			Message: "登录方式错误",
		}, fmt.Errorf("login err: 登录方式错误")
	}
}

func (l *LoginLogic) LoginByPassword(req *types.LoginRequest) (resp *types.Response, err error) {
	if req.Password == "" {
		return &types.Response{
			Code:    400,
			Message: "密码不能为空",
		}, nil
	}

	res, err := l.svcCtx.AuthRpc.Login(l.ctx, &auth.LoginRequest{
		Type:     auth.LoginType_PASSWORD,
		Email:    req.Email,
		Password: req.Password,
	})

	var data *auth.LoginResponse
	if res.Data != nil {
		data = &auth.LoginResponse{}
		err = anypb.UnmarshalTo(res.Data, data, proto.UnmarshalOptions{})
		if err != nil {
			l.Logger.Errorf("Login 消息类型转换失败：err %v")

			return &types.Response{
				Code:    400,
				Message: "消息类型转换失败",
			}, nil
		}
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}

func (l *LoginLogic) LoginByEmailCode(req *types.LoginRequest) (resp *types.Response, err error) {
	if req.EmailCode == "" {
		return &types.Response{
			Code:    400,
			Message: "验证码不能为空",
		}, nil
	}

	res, _ := l.svcCtx.AuthRpc.Login(l.ctx, &auth.LoginRequest{
		Type:      auth.LoginType_EMAIL_CODE,
		Email:     req.Email,
		EmailCode: req.EmailCode,
	})

	var data *auth.LoginResponse
	if res.Data != nil {
		data = &auth.LoginResponse{}
		_ = anypb.UnmarshalTo(res.Data, data, proto.UnmarshalOptions{})
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
