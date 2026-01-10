package logic

import (
	"context"
	"database/sql"
	"github.com/luyb177/XiaoAnBackend/auth/internal/middleware"
	"strconv"
	"time"

	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	authcode "github.com/luyb177/XiaoAnBackend/auth/pkg/code"
	"github.com/luyb177/XiaoAnBackend/auth/pkg/retry"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"
)

type GenerateInviteCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	UserDao       model.UserModel
	InviteCodeDao model.InviteCodeModel
}

func NewGenerateInviteCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateInviteCodeLogic {
	return &GenerateInviteCodeLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		UserDao:       model.NewUserModel(svcCtx.Mysql),
		InviteCodeDao: model.NewInviteCodeModel(svcCtx.Mysql),
	}
}

// GenerateInviteCode 生成邀请码
func (l *GenerateInviteCodeLogic) GenerateInviteCode(in *v1.GenerateInviteCodeRequest) (*v1.Response, error) {
	creator := middleware.MustGetUser(l.ctx)
	if creator.UID == InvalidUserID || creator.Role == "" || creator.Status != UserStatusNormal {
		l.Logger.Errorf("GenerateInviteCode err 用户未登录或登录状态异常")

		return &v1.Response{
			Code:    400,
			Message: "用户未登录或登录状态异常",
		}, nil
	}

	if in.Department == "" {
		l.Logger.Errorf("GenerateInviteCode err 部门为空")

		return &v1.Response{
			Code:    400,
			Message: "部门为空",
		}, nil
	}

	if in.TargetRole == "" {
		l.Logger.Errorf("GenerateInviteCode err 邀请码目标角色为空")

		return &v1.Response{
			Code:    400,
			Message: "邀请码目标角色为空",
		}, nil
	}

	if in.MaxUses < 0 {
		in.MaxUses = 1
	}
	if in.ExpiresAt <= 0 {
		in.ExpiresAt = 604800 // 7 day
	}

	// 验证生成者身份
	if creator.Role != SUPERADMIN && creator.Role != CLASSADMIN {
		l.Logger.Errorf("GenerateInviteCode err 用户权限不足")
		return &v1.Response{
			Code:    400,
			Message: "用户权限不足",
		}, nil
	}
	if creator.Role == CLASSADMIN {
		// 创建的是 student
		if in.TargetRole != STUDENT {
			l.Logger.Errorf("GenerateInviteCode err 用户权限不足")
			return &v1.Response{
				Code:    400,
				Message: "用户权限不足",
			}, nil
		}
	}

	// 创建邀请码
	// 指数退避
	// 失败后-可能是邀请码冲突，重新生成
	var code model.InviteCode
	var fn func() error

	fn = func() error {
		now := time.Now()
		code.Code = authcode.InviteCode()
		code.CreatorId = int64(creator.UID)
		code.CreatorName = sql.NullString{String: in.CreatorName, Valid: true}
		code.Department = sql.NullString{String: in.Department, Valid: true}
		code.MaxUses = in.MaxUses
		code.UsedCount = 0
		code.IsActive = InviteCodeActive // 有效
		code.Remark = sql.NullString{String: in.Remark, Valid: true}
		code.CreatedAt = now
		code.UpdatedAt = now
		code.ExpiresAt = sql.NullTime{Time: now.Add(time.Duration(in.ExpiresAt) * time.Second), Valid: true}
		code.TargetRole = in.TargetRole
		code.ClassId = 0
		code.Type = in.TargetRole

		_, err := l.InviteCodeDao.Insert(l.ctx, &code)
		return err
	}
	err := retry.ExponentialBackoffRetry(5, 50*time.Millisecond, time.Second, fn)

	if err != nil {
		l.Logger.Errorf("GenerateInviteCode err 生成邀请码失败")

		return &v1.Response{
			Code:    400,
			Message: "生成邀请码失败",
		}, nil
	}

	// 构建返回体
	res := v1.GenerateInviteCodeResponse{
		Code:        code.Code,
		CreatorId:   strconv.FormatUint(creator.UID, 10),
		CreatorName: in.CreatorName,
		Department:  in.Department,
		MaxUses:     in.MaxUses,
		Remark:      in.Remark,
		CreatedAt:   code.CreatedAt.Unix(),
		UpdatedAt:   code.UpdatedAt.Unix(),
		ExpiresAt:   code.ExpiresAt.Time.Unix(),
		TargetRole:  code.TargetRole,
		ClassId:     in.ClassId,
	}

	resAny, err := anypb.New(&res)
	if err != nil {
		l.Logger.Errorf("GenerateInviteCode err 消息类型转换失败")

		return &v1.Response{
			Code:    400,
			Message: "消息类型转换失败",
		}, nil
	}

	return &v1.Response{
		Code:    200,
		Message: "生成邀请码成功",
		Data:    resAny,
	}, nil
}
