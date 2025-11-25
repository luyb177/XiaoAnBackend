package logic

import (
	"context"
	"fmt"
	"sync"

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
	creatorId := l.ctx.Value("user_id").(uint64)
	creatorRole := l.ctx.Value("user_role").(string)
	creatorStatus := l.ctx.Value("user_status").(int64)

	if creatorId == 0 || creatorRole == "" || creatorStatus != 1 {
		return &v1.Response{
			Code:    400,
			Message: "请先登录",
		}, fmt.Errorf("请先登录")
	}

	// 异步一下
	var wg sync.WaitGroup
	var inviteCodes []*model.InviteCode // 用来存储查询结果
	var totalCount int64                // 用来存储邀请码总数
	var queryErr error

	// 执行 FindByCreatorId 查询
	wg.Add(1)
	go func() {
		defer wg.Done()
		inviteCodes, queryErr = l.InviteCodeDao.FindByCreatorId(l.ctx, uint64(creatorId), in.Page, in.PageSize)
		if queryErr != nil {
			logx.Errorf("获取邀请码失败: %v", queryErr)
		}
	}()

	// 执行 CountByCreatorId 查询
	wg.Add(1)
	go func() {
		defer wg.Done()
		totalCount, queryErr = l.InviteCodeDao.CountByCreatorId(l.ctx, uint64(creatorId))
		if queryErr != nil {
			logx.Errorf("获取邀请码总数失败: %v", queryErr)
		}
	}()

	// 等待所有查询任务完成
	wg.Wait()

	if queryErr != nil {
		return &v1.Response{
			Code:    400,
			Message: "获取邀请码数据失败",
		}, queryErr
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
			CreatedAt:   code.CreatedAt,
			ExpiresAt:   code.ExpiresAt.Int64,
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
