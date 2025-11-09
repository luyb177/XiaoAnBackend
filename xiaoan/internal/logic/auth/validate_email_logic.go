package auth

import (
	"context"

	auth "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateEmailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewValidateEmailLogic 验证邮箱验证码
func NewValidateEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateEmailLogic {
	return &ValidateEmailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ValidateEmailLogic) ValidateEmail(req *types.ValidateEmailRequest) (resp *types.Response, err error) {
	if req.Email == "" || req.Code == "" {
		return &types.Response{
			Code:    400,
			Message: "参数错误",
		}, nil
	}
	res, err := l.svcCtx.AuthRpc.ValidateEmailCode(l.ctx, &auth.ValidateEmailRequest{
		Email: req.Email,
		Code:  req.Code,
	})

	if err != nil {
		if res == nil {
			return &types.Response{
				Code:    400,
				Message: "验证码错误",
			}, err
		}
		return &types.Response{
			Code:    res.Code,
			Message: res.Message,
			Data:    res.Data,
		}, err
	}

	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    res.Data,
	}, nil
}
