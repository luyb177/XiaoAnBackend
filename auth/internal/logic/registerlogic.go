package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/auth/utils"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/protobuf/types/known/anypb"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	NamePrifx = "小安用户"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	UserDao    model.UserModel
	InviteCode model.InviteCodeModel
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:        ctx,
		svcCtx:     svcCtx,
		Logger:     logx.WithContext(ctx),
		UserDao:    model.NewUserModel(svcCtx.Mysql),
		InviteCode: model.NewInviteCodeModel(svcCtx.Mysql),
	}
}

// Register 注册登录
func (l *RegisterLogic) Register(in *v1.RegisterRequest) (*v1.Response, error) {
	if in.Email == "" {
		l.Logger.Errorf("Register err: 邮箱不能为空")

		return &v1.Response{
			Code:    400,
			Message: "邮箱不能为空",
		}, nil
	}

	// 先查询是否已注册
	_, err := l.UserDao.FindOneByEmail(l.ctx, in.Email)
	if err == nil || errors.Is(err, sqlx.ErrNotFound) {
		l.Logger.Errorf("Register err: %s 邮箱已注册或者邮箱错误", in.Email)

		return &v1.Response{
			Code:    400,
			Message: "邮箱已注册或出现错误",
		}, nil
	}

	// 验证邀请码
	if in.InviteCodeUsed == "" {
		l.Logger.Errorf("Register err: 邀请码为空")

		return &v1.Response{
			Code:    400,
			Message: "邀请码为空",
		}, nil
	}

	code, err := l.InviteCode.FindOneByCode(l.ctx, in.InviteCodeUsed)
	if err != nil {
		l.Logger.Errorf("Register err: 邀请码不存在")

		return &v1.Response{
			Code:    400,
			Message: "邀请码不存在",
		}, nil
	}

	// 1. 验证邀请码是否失效
	if code.IsActive != 1 {
		l.Logger.Errorf("Register err: 邀请码已失效")

		return &v1.Response{
			Code:    400,
			Message: "邀请码已失效",
		}, nil
	}

	// 2. 验证邀请码是否已使用完
	if code.UsedCount > code.MaxUses {
		l.Logger.Errorf("Register err: 邀请码已使用完")

		return &v1.Response{
			Code:    400,
			Message: "邀请码已使用完",
		}, nil
	}

	// 3. 验证其他必填信息
	if in.EmailCode == "" {
		l.Logger.Errorf("Register err: 验证码不能为空")

		return &v1.Response{
			Code:    400,
			Message: "验证码不能为空",
			Data:    nil,
		}, nil
	}
	if in.Password == "" {
		l.Logger.Errorf("Register err: 密码不能为空")

		return &v1.Response{
			Code:    400,
			Message: "密码不能为空",
			Data:    nil,
		}, nil
	}

	// 验证邮箱验证码
	getCode, err := l.svcCtx.RedisRepo.GetEmailCode(in.Email)
	if err != nil {
		l.Logger.Errorf("Register err: 邮箱[%s]获取验证码失败", in.Email)

		return &v1.Response{
			Code:    400,
			Message: fmt.Sprintf("邮箱[%s]获取验证码失败", in.Email),
		}, err
	}

	if getCode != in.EmailCode {
		l.Logger.Errorf("Register err: 邮箱[%s]验证码错误", in.Email)

		return &v1.Response{
			Code:    400,
			Message: fmt.Sprintf("邮箱[%s]验证码错误", in.Email),
		}, nil
	}

	hashPassword, err := utils.HashPassword(in.Password)
	if err != nil {
		l.Logger.Errorf("Register err: 密码加密失败")

		return &v1.Response{
			Code:    400,
			Message: "密码加密失败",
		}, nil
	}

	// 创建用户
	user := model.User{
		Name:           NamePrifx + in.InviteCodeUsed,
		Email:          in.Email,
		Password:       hashPassword,
		Department:     code.Department,
		Role:           code.TargetRole,
		ClassId:        uint64(code.ClassId),
		Status:         1, // 1 正常
		InviteCodeUsed: sql.NullString{String: code.Code, Valid: true},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 事务
	err = l.svcCtx.Mysql.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 新增用户
		_, err = l.UserDao.InsertWithSession(ctx, session, &user)
		if err != nil {
			return err
		}

		// 修改邀请码
		code.UsedCount++
		if code.UsedCount >= code.MaxUses {
			code.IsActive = 0
		}
		return l.InviteCode.UpdateWithSession(ctx, session, code)
	})

	if err != nil {
		l.Logger.Errorf("Register err: 注册用户失败，请稍后尝试,%v", err)
		return &v1.Response{
			Code:    400,
			Message: "注册用户失败，请稍后尝试",
		}, nil
	}

	// 构造返回内容
	res := &v1.RegisterResponse{
		Id:             strconv.FormatUint(user.Id, 10),
		Name:           user.Name,
		Email:          user.Email,
		Avatar:         user.Avatar.String,
		Phone:          user.Phone.String,
		Department:     user.Department.String,
		Role:           user.Role,
		ClassId:        user.ClassId,
		Status:         user.Status,
		InviteCodeUsed: user.InviteCodeUsed.String,
		CreatedAt:      user.CreatedAt.Unix(),
		UpdatedAt:      user.UpdatedAt.Unix(),
	}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Logger.Errorf("Register err: 消息类型转换失败")

		return &v1.Response{
			Code:    400,
			Message: "消息类型转换失败",
		}, nil
	}

	return &v1.Response{
		Code:    200,
		Message: "注册成功",
		Data:    resAny,
	}, nil
}
