package logic

import (
	"context"
	"fmt"
	"github.com/luyb177/XiaoAnBackend/auth/internal/model"

	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	UserDao model.UserModel
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:     ctx,
		svcCtx:  svcCtx,
		Logger:  logx.WithContext(ctx),
		UserDao: model.NewUserModel(svcCtx.Mysql),
	}
}

// Register 注册登录
func (l *RegisterLogic) Register(in *v1.RegisterRequest) (*v1.Response, error) {
	if in.Email == "" || in.EmailCode == "" {
		return &v1.Response{
			Code:    400,
			Message: "邮箱或验证码不能为空",
			Data:    nil,
		}, fmt.Errorf("邮箱或验证码不能为空")
	}
	if in.Password == "" {
		return &v1.Response{
			Code:    400,
			Message: "密码不能为空",
			Data:    nil,
		}, fmt.Errorf("密码不能为空")
	}
	if in.Name == "" {
		return &v1.Response{
			Code:    400,
			Message: "用户名不能为空",
			Data:    nil,
		}, fmt.Errorf("用户名不能为空")
	}
	if in.Phone == "" {
		return &v1.Response{
			Code:    400,
			Message: "手机号不能为空",
			Data:    nil,
		}, fmt.Errorf("手机号不能为空")
	}
	if in.Department == "" {
		return &v1.Response{
			Code:    400,
			Message: "部门不能为空",
			Data:    nil,
		}, fmt.Errorf("部门不能为空")
	}

	if in.Role == "" {
		return &v1.Response{
			Code:    400,
			Message: "角色为空",
		}, fmt.Errorf("角色为空")
	}

	// 验证邮箱验证码
	getCode, err := l.svcCtx.RedisRepo.GetEmailCode(in.Email)
	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: fmt.Sprintf("邮箱[%s]获取验证码失败", in.Email),
		}, err
	}

	if getCode != in.EmailCode {
		return &v1.Response{
			Code:    400,
			Message: fmt.Sprintf("邮箱[%s]验证码错误", in.Email),
		}, fmt.Errorf("验证码错误")
	}

	// 验证 邀请码

	return &v1.Response{}, nil
}
