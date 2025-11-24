package logic

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/auth/utils"
	"google.golang.org/protobuf/types/known/anypb"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
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
	if in.CreatorId == "" || in.CreatorName == "" {
		return &v1.Response{
			Code:    400,
			Message: "生成人为空",
		}, fmt.Errorf("生成人为空")
	}
	if in.Department == "" {
		return &v1.Response{
			Code:    400,
			Message: "部门为空",
		}, fmt.Errorf("部门为空")
	}

	if in.TargetRole == "" {
		return &v1.Response{
			Code:    400,
			Message: "邀请码目标角色为空",
		}, fmt.Errorf("邀请码目标角色为空")
	}

	if in.MaxUses < 0 {
		in.MaxUses = 1
	}
	if in.ExpiresAt == 0 {
		in.ExpiresAt = 604800 // 7 day
	}

	// 验证生成者身份 // Todo 此时有 token 了所以可以将 token 解析

	// 创建邀请码
	// 指数退避
	// 失败后-可能是邀请码冲突，重新生成
	var code model.InviteCode
	var fn func() error
	fn = func() error {
		// 生成邀请码
		creatorIdInt, err := strconv.Atoi(in.CreatorId)
		if err != nil {
			return fmt.Errorf("生成者ID转换失败,err: %v", err)
		}
		code.Code = utils.GenerateInviteCode()
		code.CreatorId = int64(creatorIdInt)
		code.CreatorName = sql.NullString{String: in.CreatorName, Valid: true}
		code.Department = sql.NullString{String: in.Department, Valid: true}
		code.MaxUses = in.MaxUses
		code.UsedCount = 0
		code.IsActive = 1 // 有效
		code.Remark = sql.NullString{String: in.Remark, Valid: true}
		code.CreatedAt = time.Now().Unix()
		code.ExpiresAt = sql.NullInt64{Int64: in.ExpiresAt, Valid: true}
		code.TargetRole = in.TargetRole
		code.ClassId = 0
		code.Type = in.TargetRole

		_, err = l.InviteCodeDao.Insert(l.ctx, &code)
		return err
	}
	err := utils.ExponentialBackoffRetry(5, 50*time.Millisecond, time.Second, fn)

	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: "生成邀请码失败",
		}, fmt.Errorf("生成邀请码失败,err: %v", err)
	}

	// 构建返回体
	res := v1.GenerateInviteCodeResponse{
		Code:        code.Code,
		CreatorId:   in.CreatorId,
		CreatorName: in.CreatorName,
		Department:  in.Department,
		MaxUses:     in.MaxUses,
		Remark:      in.Remark,
		CreatedAt:   code.CreatedAt,
		ExpiresAt:   code.ExpiresAt.Int64,
		TargetRole:  code.TargetRole,
		ClassId:     in.ClassId,
	}

	resAny, err := anypb.New(&res)
	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: "消息类型转换失败",
		}, fmt.Errorf("消息类型转换失败")
	}

	return &v1.Response{
		Code:    200,
		Message: "生成邀请码成功",
		Data:    resAny,
	}, nil
}
