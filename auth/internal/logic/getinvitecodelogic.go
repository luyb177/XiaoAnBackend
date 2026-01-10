package logic

import (
	"context"
	"fmt"
	"sync"

	"github.com/luyb177/XiaoAnBackend/auth/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"
)

type GetInviteCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	UserDao       model.UserModel
	InviteCodeDao model.InviteCodeModel
}

func NewGetInviteCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInviteCodeLogic {
	return &GetInviteCodeLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		UserDao:       model.NewUserModel(svcCtx.Mysql),
		InviteCodeDao: model.NewInviteCodeModel(svcCtx.Mysql),
	}
}

func (l *GetInviteCodeLogic) GetInviteCode(in *v1.GetInviteCodeRequest) (*v1.Response, error) {
	creator := middleware.MustGetUser(l.ctx)
	if creator.UID == 0 || creator.Role == "" || creator.Status != 1 {
		l.Logger.Errorf("GenerateInviteCode err 用户未登录或登录状态异常")

		return &v1.Response{
			Code:    400,
			Message: "用户未登录或登录状态异常",
		}, nil
	}

	// 异步一下
	var wg sync.WaitGroup
	var inviteCodes []*model.InviteCode // 用来存储查询结果
	var totalCount int64                // 用来存储邀请码总数
	var findErr error
	var countErr error

	// 执行 FindByCreatorId 查询
	wg.Add(1)
	go func() {
		defer wg.Done()
		inviteCodes, findErr = l.InviteCodeDao.FindByCreatorId(l.ctx, creator.UID, in.Page, in.PageSize)
		if findErr != nil {
			l.Logger.Errorf("获取邀请码失败: %v", findErr)
			return
		}
	}()

	// 执行 CountByCreatorId 查询
	wg.Add(1)
	go func() {
		defer wg.Done()
		totalCount, countErr = l.InviteCodeDao.CountByCreatorId(l.ctx, creator.UID)
		if countErr != nil {
			l.Logger.Errorf("获取邀请码总数失败: %v", countErr)
			return
		}
	}()

	// 等待所有查询任务完成
	wg.Wait()

	if findErr != nil || countErr != nil {
		return &v1.Response{
			Code:    400,
			Message: "获取邀请码数据失败",
		}, nil
	}

	var codes []*v1.InviteCode
	for _, code := range inviteCodes {
		codes = append(codes, &v1.InviteCode{
			Code:        code.Code,
			CreatorId:   fmt.Sprintf("%d", code.CreatorId),
			CreatorName: code.CreatorName.String,
			Department:  code.Department.String,
			MaxUses:     code.MaxUses,
			UsedCount:   code.UsedCount,
			IsActive:    code.IsActive,
			Remark:      code.Remark.String,
			CreatedAt:   code.CreatedAt.Unix(),
			UpdatedAt:   code.UpdatedAt.Unix(),
			ExpiresAt:   code.ExpiresAt.Time.Unix(),
		})
	}

	// 构造响应
	responsePb := &v1.GetInviteCodeResponse{
		Codes:    codes,
		Total:    totalCount,
		Page:     in.Page,
		PageSize: in.PageSize,
	}

	// 将响应结构转换为 anypb.Any
	anyResponse, err := anypb.New(responsePb)
	if err != nil {
		return nil, err
	}

	return &v1.Response{
		Code:    200,
		Message: "获取邀请码成功",
		Data:    anyResponse,
	}, nil
}
