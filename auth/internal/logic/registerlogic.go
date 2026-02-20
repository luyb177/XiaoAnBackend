package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/anypb"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	v1 "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/auth/pkg/email"
	"github.com/luyb177/XiaoAnBackend/auth/pkg/password"
	"github.com/luyb177/XiaoAnBackend/auth/pkg/taskqueue/tasks"
	"github.com/luyb177/XiaoAnBackend/infra/constants"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	UserDao    model.UserModel
	InviteCode model.InviteCodeModel
	ClassDao   model.ClassModel
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:        ctx,
		svcCtx:     svcCtx,
		Logger:     logx.WithContext(ctx),
		UserDao:    model.NewUserModel(svcCtx.Mysql),
		InviteCode: model.NewInviteCodeModel(svcCtx.Mysql),
		ClassDao:   model.NewClassModel(svcCtx.Mysql),
	}
}

// Register 注册登录
func (l *RegisterLogic) Register(in *v1.RegisterRequest) (*v1.Response, error) {
	if resp := l.validate(in); resp != nil {
		return resp, nil
	}

	// email 基本验证
	// 	1. 长度
	if len(in.Email) > 254 {
		return bad("邮箱长度不能超过254个字符"), nil
	}

	// 	2. trim spaces and to lower
	in.Email = email.CanonicalEmail(in.Email)

	// 	3. 基本格式验证
	if !email.IsValidEmail(in.Email) {
		return bad("邮箱格式不正确"), nil
	}

	// 使用 errGroup 并发校验 email code 和 invite code
	var (
		emailCode string
		emailOk   bool
		code      *model.InviteCode
	)

	g, ctx := errgroup.WithContext(l.ctx)

	g.Go(func() error {
		var err error
		emailCode, emailOk, err = l.svcCtx.RedisRepo.EmailRepo.GetEmailCode(in.Email)
		return err
	})

	g.Go(func() error {
		var err error
		code, err = l.InviteCode.FindUsableByCode(ctx, in.InviteCodeUsed)
		if err != nil {
			return err
		}

		if code.ClassId != 0 {
			if code.TargetRole != constants.STUDENT {
				return errors.New("邀请码关联的角色不合法")
			}

			_, err = l.ClassDao.FindNormalOne(ctx, code.ClassId)
			if err != nil {
				if errors.Is(err, model.ErrNotFound) {
					return errors.New("邀请码关联的班级不存在")
				}
				l.Errorf("查询班级失败, classId=%d, err=%v", code.ClassId, err)
				return errors.New("系统繁忙，请稍后尝试")
			}
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return bad("邀请码不存在"), nil
		}
		l.Errorf("Register err: 并发校验失败,%v", err)
		return internal("系统繁忙，请稍后尝试"), nil
	}

	// 串行校验
	if !emailOk {
		return bad("邮箱验证码已过期，请重新获取"), nil
	}
	if emailCode != in.EmailCode {
		return bad("邮箱验证码错误"), nil
	}

	// 查询用户是否存在
	_, err := l.UserDao.FindOneByEmailWithNotDelete(l.ctx, in.Email)
	if err == nil {
		return bad("该邮箱已注册"), nil
	} else if !errors.Is(err, model.ErrNotFound) {
		l.Errorf("Register err: 查询用户失败,%v", err)
		return internal("查询用户失败"), nil
	}

	hashPassword, err := password.Hash(in.Password)
	if err != nil {
		l.Errorf("Register err: 密码加密失败, email=%s, err=%v", in.Email, err)
		return bad("密码存在安全问题，请更换密码"), nil
	}

	// 创建用户
	now := time.Now()
	user := model.User{
		Name:           NamePrefix + rand.String(8),
		Email:          in.Email,
		Password:       hashPassword,
		Department:     code.Department,
		Role:           code.TargetRole,
		ClassId:        code.ClassId,
		Status:         constants.UserStatusNormal, // 1 正常
		InviteCodeUsed: sql.NullString{String: code.Code, Valid: true},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// 事务
	err = l.svcCtx.Mysql.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 修改邀请码
		result, err := l.InviteCode.IncrUsedCountWithSession(ctx, session, code.Id)
		if err != nil {
			return err
		}
		affect, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affect == 0 {
			return errors.New("邀请码已使用完或失效")
		}

		// 新增用户
		result, err = l.UserDao.InsertWithSession(ctx, session, &user)
		if err != nil {
			return err
		}
		userID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		user.Id = uint64(userID)

		// 班级成员数+1
		if code.ClassId != 0 {
			result, err = l.ClassDao.IncrStudentCountWithSession(ctx, session, code.ClassId)
			if err != nil {
				return err
			}
			affect, err = result.RowsAffected()
			if err != nil {
				return err
			}
			if affect == 0 {
				return errors.New("加入班级失败，班级不存在或已失效")
			}
		}
		return nil
	})

	if err != nil {
		l.Errorf("Register err: %s注册用户失败，请稍后尝试,%v", in.Email, err)
		return bad("注册用户失败，请稍后尝试"), nil
	}

	// 幕后小工作
	emailRelationTask := &tasks.EmailRelationTask{
		Type: tasks.EmailCodeRelationDelete,
		To:   in.Email,
		Code: in.EmailCode,
	}

	if err := l.svcCtx.TaskQueue.Enqueue(l.ctx, emailRelationTask); err != nil {
		l.Errorf("Register 入队列失败,%v", err)
	}

	// 构造返回内容
	res := &v1.RegisterResponse{User: &v1.User{
		Id:             user.Id,
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
	}}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("Register err: 消息类型转换失败,%v", err)
		return internal("系统内部错误"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "注册成功",
		Data:    resAny,
	}, nil
}

func (l *RegisterLogic) validate(in *v1.RegisterRequest) *v1.Response {
	switch {
	case in.Email == "":
		return bad("邮箱不能为空")
	case in.EmailCode == "":
		return bad("验证码不能为空")
	case in.Password == "":
		return bad("密码不能为空")
	case in.InviteCodeUsed == "":
		return bad("邀请码不能为空")
	}
	return nil
}
