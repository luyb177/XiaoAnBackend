package logic

import (
	"context"
	"database/sql"
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
		ctx:     ctx,
		svcCtx:  svcCtx,
		Logger:  logx.WithContext(ctx),
		UserDao: model.NewUserModel(svcCtx.Mysql),
	}
}

// Register 注册登录
func (l *RegisterLogic) Register(in *v1.RegisterRequest) (*v1.Response, error) {
	// 验证邀请码
	if in.InviteCodeUsed == "" {
		return &v1.Response{
			Code:    400,
			Message: "邀请码为空",
		}, fmt.Errorf("邀请码为空")
	}

	code, err := l.InviteCode.FindOneByCode(l.ctx, in.InviteCodeUsed)
	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: "邀请码不存在",
		}, fmt.Errorf("邀请码不存在")
	}

	if code.IsActive != 1 {
		return &v1.Response{
			Code:    400,
			Message: "邀请码已失效",
		}, fmt.Errorf("邀请码已失效")
	}

	if code.UsedCount > code.MaxUses {
		return &v1.Response{
			Code:    400,
			Message: "邀请码已使用完",
		}, fmt.Errorf("邀请码已使用完")
	}

	// 验证其他必填信息
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

	hashPassword, err := utils.HashPassword(in.Password)
	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: "密码加密失败",
		}, err
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
		CreatedAt:      time.Now().Unix(),
		UpdatedAt:      time.Now().Unix(),
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
		return l.InviteCode.UpdateWithSession(ctx, session, code)
	})

	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: "注册用户失败，请稍后尝试",
		}, fmt.Errorf("注册用户失败，请稍后尝试 err: %v", err)
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
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: "消息类型转换失败",
		}, fmt.Errorf("消息类型转换失败")
	}

	return &v1.Response{
		Code:    200,
		Message: "注册成功",
		Data:    resAny,
	}, nil
}
