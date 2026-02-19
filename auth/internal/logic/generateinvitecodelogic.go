package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	v1 "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	authcode "github.com/luyb177/XiaoAnBackend/auth/pkg/code"
	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"
	"github.com/luyb177/XiaoAnBackend/infra/retry"
)

const (
	defaultExpireSeconds = 24 * 3600      // 默认过期时间24小时
	maxExpireSeconds     = 30 * 24 * 3600 // 30天
)

type GenerateInviteCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	UserDao       model.UserModel
	InviteCodeDao model.InviteCodeModel
	ClassDao      model.ClassModel
}

func NewGenerateInviteCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateInviteCodeLogic {
	return &GenerateInviteCodeLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		UserDao:       model.NewUserModel(svcCtx.Mysql),
		InviteCodeDao: model.NewInviteCodeModel(svcCtx.Mysql),
		ClassDao:      model.NewClassModel(svcCtx.Mysql),
	}
}

// GenerateInviteCode 生成邀请码
func (l *GenerateInviteCodeLogic) GenerateInviteCode(in *v1.GenerateInviteCodeRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || user.Role == "" || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或登录状态异常"), nil
	}

	if resp := l.validate(in); resp != nil {
		return resp, nil
	}

	if in.MaxUses == 0 {
		in.MaxUses = 1
	}
	if in.ExpiresAt <= 0 {
		in.ExpiresAt = defaultExpireSeconds // 默认24小时过期
	}
	if in.ExpiresAt > maxExpireSeconds {
		return bad("过期时间不能超过30天"), nil
	}

	// 验证用户信息
	dbUser, err := l.UserDao.FindOneWithNotDelete(l.ctx, user.UID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return bad("用户不存在"), nil
		}
		l.Errorf("FindOneWithNotDelete err: %v", err)
		return bad("用户信息查询失败"), nil
	}

	// 验证班级
	if in.ClassId != 0 {
		if in.TargetRole != constants.STUDENT {
			return bad("只有学生邀请码可以指定班级"), nil
		}
		// 验证班级是否存在
		class, err := l.ClassDao.FindNormalOne(l.ctx, in.ClassId)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return bad("班级不存在"), nil
			}
			l.Errorf("FindNormalOne err: %v", err)
			return internal("系统繁忙，请稍后再试"), nil
		}
		if class.AdminId != user.UID {
			if dbUser.Role != constants.SUPERADMIN && dbUser.Role != constants.STAFF {
				return bad("没有权限生成该班级的邀请码"), nil
			}
		}
	}

	// 验证身份 - 权限检查
	if !hasPermission(dbUser.Role, in.TargetRole) {
		return bad("没有权限生成该角色的邀请码"), nil
	}

	// 生成邀请码
	var code model.InviteCode
	fn := func() error {
		now := time.Now()
		code.Code = authcode.InviteCode()
		code.CreatorId = user.UID
		code.Department = sql.NullString{String: in.Department, Valid: true}
		code.MaxUses = in.MaxUses
		code.UsedCount = 0
		code.IsActive = InviteCodeActive // 有效
		code.Remark = sql.NullString{String: in.Remark, Valid: true}
		code.CreatedAt = now
		code.UpdatedAt = now
		code.ExpiresAt = sql.NullTime{Time: now.Add(time.Duration(in.ExpiresAt) * time.Second), Valid: true}
		code.TargetRole = in.TargetRole
		code.ClassId = in.ClassId

		_, err := l.InviteCodeDao.Insert(l.ctx, &code)
		return err
	}

	err = retry.Do(
		l.ctx,
		fn,
		retry.WithJitter(),
		retry.WithRetryIf(func(err error) bool {
			return errors.Is(err, model.ErrDuplicateEntry)
		}),
		retry.WithOnRetry(func(attempt int, err error, nextDelay time.Duration) {
			l.Errorf("GenerateInviteCode Insert %d failed: %v, next retry in %v", attempt, err, nextDelay)
		}),
	)

	if err != nil {
		return bad("生成邀请码失败"), nil
	}

	res := &v1.GenerateInviteCodeResponse{
		Code: &v1.InviteCode{
			Code:       code.Code,
			CreatorId:  code.CreatorId,
			Department: code.Department.String,
			MaxUses:    code.MaxUses,
			UsedCount:  code.UsedCount,
			Remark:     code.Remark.String,
			CreatedAt:  code.CreatedAt.Unix(),
			ExpiresAt:  code.ExpiresAt.Time.Unix(),
			TargetRole: code.TargetRole,
			ClassId:    code.ClassId,
			UpdatedAt:  code.UpdatedAt.Unix(),
		},
	}

	resAny, err := anypb.New(res)
	if err != nil {
		return internal("系统繁忙，请稍后再试"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "生成邀请码成功",
		Data:    resAny,
	}, nil
}

func (l *GenerateInviteCodeLogic) validate(in *v1.GenerateInviteCodeRequest) *v1.Response {
	switch {
	case in.Department == "":
		return bad("部门为空")
	case in.TargetRole == "":
		return bad("邀请码目标角色为空")
	}
	return nil
}

func hasPermission(userRole, targetRole string) bool {
	switch userRole {
	case constants.SUPERADMIN:
		return targetRole == constants.SUPERADMIN || targetRole == constants.STAFF || targetRole == constants.CLASSADMIN || targetRole == constants.STUDENT
	case constants.STAFF:
		return targetRole == constants.STAFF || targetRole == constants.CLASSADMIN || targetRole == constants.STUDENT
	case constants.CLASSADMIN:
		return targetRole == constants.STUDENT
	default:
		return false
	}
}
