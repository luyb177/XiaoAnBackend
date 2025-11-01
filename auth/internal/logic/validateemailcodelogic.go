package logic

import (
	"context"
	"fmt"

	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateEmailCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateEmailCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateEmailCodeLogic {
	return &ValidateEmailCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ValidateEmailCodeLogic) ValidateEmailCode(in *v1.ValidateEmailRequest) (*v1.Response, error) {
	if in.Email == "" {
		return &v1.Response{
			Code:    400,
			Message: "邮箱不能为空",
		}, fmt.Errorf("邮箱不能为空")
	}
	if in.Code == "" {
		return &v1.Response{
			Code:    400,
			Message: "验证码不能为空",
			Data:    nil,
		}, fmt.Errorf("验证码不能为空")
	}

	getCode, err := l.svcCtx.RedisRepo.GetEmailCode(in.Email)
	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: fmt.Sprintf("邮箱[%s]获取验证码失败", in.Email),
		}, err
	}

	if getCode != in.Code {
		return &v1.Response{
			Code:    400,
			Message: fmt.Sprintf("邮箱[%s]验证码错误", in.Email),
		}, fmt.Errorf("验证码错误")
	}

	// 删除验证码
	err = l.svcCtx.RedisRepo.DelEmailCode(in.Email)
	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: fmt.Sprintf("邮箱[%s]删除验证码失败", in.Email),
		}, err
	}

	return &v1.Response{
		Code:    200,
		Message: fmt.Sprintf("邮箱[%s]验证成功", in.Email),
		Data:    nil,
	}, nil
}
