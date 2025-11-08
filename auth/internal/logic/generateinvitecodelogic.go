package logic

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	"github.com/luyb177/XiaoAnBackend/auth/utils"
	"strconv"
	"time"

	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"

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

// GenerateInviteCode 邀请码
func (l *GenerateInviteCodeLogic) GenerateInviteCode(in *v1.GenerateInviteCodeRequest) (*v1.Response, error) {
	if in.Count <= 0 {
		return &v1.Response{
			Code:    400,
			Message: "请输入正确的数量",
		}, fmt.Errorf("请输入正确的数量")
	}
	if in.CreatorId == "" || in.CreatorName == "" {
		return &v1.Response{
			Code:    400,
			Message: "请输入正确的创建人信息",
		}, fmt.Errorf("请输入正确的创建人信息")
	}
	if in.Department == "" {
		return &v1.Response{
			Code:    400,
			Message: "请输入正确的部门信息",
		}, fmt.Errorf("请输入正确的部门信息")
	}

	// 验证创建人信息
	creatorId, err := strconv.Atoi(in.CreatorId)
	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: "请输入正确的创建人信息",
		}, fmt.Errorf("请输入正确的创建人信息")
	}
	res, err := l.UserDao.FindOne(l.ctx, uint64(creatorId))
	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: "未找到创建人信息",
		}, fmt.Errorf("未找到创建人信息")
	}

	if res.Name != in.CreatorName {
		return &v1.Response{
			Code:    400,
			Message: "创建人姓名不一致",
		}, fmt.Errorf("创建人姓名不一致")
	}

	if res.Department.Valid && res.Department.String != in.Department {
		return &v1.Response{
			Code:    400,
			Message: "创建人部门不一致",
		}, fmt.Errorf("创建人部门不一致")
	}

	// 生成邀请码
	// 有效期
	expiresAt := sql.NullTime{}
	if in.ExpiresAt > 0 {
		expiresAt = sql.NullTime{
			Time:  time.Unix(in.ExpiresAt, 0), // 将 int64 时间戳转为 time.Time
			Valid: true,                       // 表示这个值是有效的
		}
	}
	inviteCodes := make([]*model.InviteCode, 0, in.Count)
	for i := 0; i < int(in.Count); i++ {
		code := utils.GenerateInviteCode()
		inviteCodes = append(inviteCodes, &model.InviteCode{
			Code:        code,
			CreatorId:   uint64(creatorId),
			CreatorName: sql.NullString{String: in.CreatorName, Valid: true},
			Department:  sql.NullString{String: in.Department, Valid: true},
			MaxUses:     int64(in.Count),
			UsedCount:   0,
			IsActive:    1,
			Remark:      sql.NullString{String: in.Remark, Valid: in.Remark != ""},
			CreatedAt:   time.Now(),
			ExpiresAt:   expiresAt,
		})
	}

	// 插入
	go func(invites []*model.InviteCode) {
		for _, v := range invites {
			_, err = l.InviteCodeDao.Insert(l.ctx, v)
			if err != nil {
				logx.Errorf("用户%v 插入邀请码失败: %v", v.CreatorName, err)
			}
		}
	}(inviteCodes)

	return &v1.Response{
		Code:    200,
		Message: fmt.Sprintf("正在生成%v个邀请码", in.Count),
	}, nil
}
