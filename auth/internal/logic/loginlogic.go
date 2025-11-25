package logic

import (
	"context"
	"fmt"
	"github.com/luyb177/XiaoAnBackend/auth/internal/jwt"
	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	"github.com/luyb177/XiaoAnBackend/auth/utils"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	UserDao model.UserModel
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:     ctx,
		svcCtx:  svcCtx,
		Logger:  logx.WithContext(ctx),
		UserDao: model.NewUserModel(svcCtx.Mysql),
	}
}

func (l *LoginLogic) Login(in *v1.LoginRequest) (*v1.Response, error) {
	// 验证邮箱验证码
	if in.Email == "" {
		return &v1.Response{
			Code:    400,
			Message: "邮箱为空",
		}, fmt.Errorf("邮箱为空")
	}

	switch in.Type {
	case v1.LoginType_EMAIL_CODE:
		msg, flag := l.validateEmailCode(in)
		if !flag {
			return &v1.Response{
				Code:    400,
				Message: msg,
			}, fmt.Errorf(msg)
		}
	case v1.LoginType_PASSWORD:
		msg, flag := l.validatePassword(in)
		if !flag {
			return &v1.Response{
				Code:    400,
				Message: msg,
			}, fmt.Errorf(msg)
		}
	}
	// 验证成功 获取用户信息
	user, err := l.UserDao.FindOneByEmail(l.ctx, in.Email)
	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: "用户不存在",
		}, fmt.Errorf("用户不存在")
	}
	if user.Status != 0 {
		return &v1.Response{
			Code:    400,
			Message: "用户被禁用",
		}, fmt.Errorf("用户被禁用")
	}

	// 生成 token
	token, err := l.svcCtx.JWTHandler.SetJWTToken(jwt.ClaimsParams{
		UserId:     user.Id,
		UserRole:   user.Role,
		UserStatus: user.Status,
	})

	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: "生成token失败",
		}, fmt.Errorf("生成token失败 err: %v", err)
	}

	// 构造返回内容
	resUser := v1.User{
		Id:             user.Id,
		Name:           user.Name,
		Email:          user.Email,
		Avatar:         user.Avatar.String,
		Phone:          user.Phone.String,
		Password:       "",
		Department:     user.Department.String,
		Role:           user.Role,
		ClassId:        user.ClassId,
		Status:         user.Status,
		InviteCodeUsed: user.InviteCodeUsed.String,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}
	res := v1.LoginResponse{
		Token: token,
		User:  &resUser,
	}
	resAny, err := anypb.New(&res)
	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: "消息类型转换失败",
		}, fmt.Errorf("消息类型转换失败 err: %v", err)
	}

	return &v1.Response{
		Code:    200,
		Message: "登录成功",
		Data:    resAny,
	}, nil
}

func (l *LoginLogic) validateEmailCode(in *v1.LoginRequest) (msg string, flag bool) {
	if in.EmailCode == "" {
		return "邮箱验证码为空", false
	}

	// 获取邮箱验证码
	getCode, err := l.svcCtx.RedisRepo.GetEmailCode(in.Email)
	if err != nil {
		return "未获取到验证码", false
	}
	if getCode != in.EmailCode {
		return "验证码错误", false
	}
	return "", true
}

func (l *LoginLogic) validatePassword(in *v1.LoginRequest) (msg string, flag bool) {
	if in.Password == "" {
		return "密码为空", false
	}

	// 获取用户信息
	user, err := l.UserDao.FindOneByEmail(l.ctx, in.Email)
	if err != nil {
		return "用户不存在", false
	}
	if !utils.CheckPasswordHash(in.Password, user.Password) {
		return "密码错误", false
	}
	return "", true
}
