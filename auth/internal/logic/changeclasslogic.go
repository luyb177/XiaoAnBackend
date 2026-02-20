package logic

import (
	"context"
	"errors"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/sync/errgroup"

	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"
)

type ChangeClassLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	UserDao       model.UserModel
	InviteCodeDao model.InviteCodeModel
	ClassDao      model.ClassModel
}

func NewChangeClassLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeClassLogic {
	return &ChangeClassLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		UserDao:       model.NewUserModel(svcCtx.Mysql),
		InviteCodeDao: model.NewInviteCodeModel(svcCtx.Mysql),
		ClassDao:      model.NewClassModel(svcCtx.Mysql),
	}
}

// ChangeClass 切换班级
func (l *ChangeClassLogic) ChangeClass(in *v1.ChangeClassRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || user.Role == "" || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或登录状态异常"), nil
	}

	if user.Role != constants.STUDENT {
		return bad("只有学生可以切换班级"), nil
	}

	if strings.TrimSpace(in.InviteCode) == "" {
		return bad("参数错误"), nil
	}

	// 使用 errgroup 同时获取学生信息和邀请码信息，减少响应时间
	var (
		student       *model.User
		studentErr    error
		inviteCode    *model.InviteCode
		inviteCodeErr error
	)

	g, ctx := errgroup.WithContext(l.ctx)

	g.Go(func() error {
		// 获取该学生目前信息
		var err error
		student, err = l.UserDao.FindOneWithNotDelete(ctx, user.UID)
		studentErr = err
		return err
	})

	// 获取邀请码信息
	g.Go(func() error {
		var err error
		inviteCode, err = l.InviteCodeDao.FindUsableByCode(ctx, in.InviteCode)
		inviteCodeErr = err
		return err
	})

	if err := g.Wait(); err != nil {
		if errors.Is(studentErr, model.ErrNotFound) {
			return bad("用户不存在"), nil
		}
		if errors.Is(inviteCodeErr, model.ErrNotFound) {
			return bad("邀请码不存在或不可用"), nil
		}
		l.Errorf("ChangeClass err: %v", err)
		return bad("切换班级失败"), nil
	}

	if student.ClassId == inviteCode.ClassId {
		return bad("新班级不能和原班级相同"), nil
	}

	// todo 事务可能比较慢，可以考虑异步修改
	err := l.svcCtx.Mysql.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 修改邀请码
		result, err := l.InviteCodeDao.IncrUsedCountWithSession(ctx, session, inviteCode.Id)
		if err != nil {
			return err
		}
		affect, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affect == 0 {
			return errors.New("邀请码已被使用完或失效")
		}

		// 修改学生班级
		result, err = l.UserDao.SwitchUserClassWithSession(ctx, session, inviteCode.ClassId, inviteCode.Department.String, student)
		if err != nil {
			return err
		}
		affect, err = result.RowsAffected()
		if err != nil {
			return err
		}
		if affect == 0 {
			return errors.New("切换班级失败，可能是因为学生信息已过期，请刷新后重试")
		}

		// 新班级学生数量+1
		result, err = l.ClassDao.IncrStudentCountWithSession(ctx, session, inviteCode.ClassId)
		if err != nil {
			return err
		}
		affect, err = result.RowsAffected()
		if err != nil {
			return err
		}
		if affect == 0 {
			return errors.New("切换班级失败，可能是因为班级信息已过期，请刷新后重试")
		}

		// 旧班级学生数量-1
		if student.ClassId != 0 {
			result, err = l.ClassDao.DecrStudentCountWithSession(ctx, session, student.ClassId)
			if err != nil {
				return err
			}
			affect, err = result.RowsAffected()
			if err != nil {
				return err
			}
			if affect == 0 {
				return errors.New("切换班级失败，可能是因为班级信息已过期，请刷新后重试")
			}
		}
		return nil
	})

	if err != nil {
		l.Errorf("ChangeClass err: %v", err)
		return bad("切换班级失败"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "切换班级成功",
	}, nil
}
