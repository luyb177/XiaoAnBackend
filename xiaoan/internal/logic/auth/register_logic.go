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
	if req.Email == "" {
		return &types.Response{
			Code:    400,
			Message: "邮箱不能为空",
		}, nil
	}
	if req.Password == "" {
		return &types.Response{
			Code:    400,
			Message: "密码不能为空",
		}, nil
	}
	if req.EmailCode == "" {
		return &types.Response{
			Code:    400,
			Message: "验证码不能为空",
		}, nil
	}
	if req.InviteCodeUsed == "" {
		return &types.Response{
			Code:    400,
			Message: "邀请码不能为空",
		}, nil
	}
	res, err := l.svcCtx.AuthRpc.Register(l.ctx, &auth.RegisterRequest{
		Email:          req.Email,
		EmailCode:      req.EmailCode,
		Password:       req.Password,
		InviteCodeUsed: req.InviteCodeUsed,
	})
	if err != nil {
		return &types.Response{
			Code:    400,
			Message: "注册失败",
		}, err
	}

	data := &auth.RegisterResponse{} // 你的具体 Protobuf 消息对象
	fmt.Println("==", res.Data.Value, "==")
	if res.Data != nil {
		_ = anypb.UnmarshalTo(res.Data, data, proto.UnmarshalOptions{})
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    data,
	}, nil
}
