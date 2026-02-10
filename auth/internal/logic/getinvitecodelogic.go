package logic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/auth/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	v1 "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/auth/pkg/code/convert"
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
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == InvalidUserID || user.Role == "" || user.Status != UserStatusNormal {
		return bad("用户未登录或登录状态异常"), nil
	}

	if in.PageSize <= 0 {
		in.PageSize = 10
	}

	// 查询多一条记录，来判断是否有下一页
	limit := in.PageSize + 1

	var (
		list []*model.InviteCode
		err  error
	)

	if in.Cursor == 0 {
		// 首次查询
		list, err = l.InviteCodeDao.FindManyByCreatorId(l.ctx, user.UID, limit)
	} else {
		// 继续查询
		list, err = l.InviteCodeDao.FindManyByCreatorIdWithCursor(l.ctx, user.UID, in.Cursor, limit)
	}

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &v1.Response{
				Code:    404,
				Message: "没有更多了",
			}, nil
		}
		l.Errorf("GetInviteCodeLogic FindManyByCreatorIdWithCursor error: %v", err)
		return internal("系统繁忙，请稍后再试"), nil
	}

	hasMore := int64(len(list)) > in.PageSize
	if hasMore {
		list = list[:in.PageSize]
	}

	nextCursor := uint64(0)
	if len(list) > 0 {
		nextCursor = list[len(list)-1].Id
	}

	inviteCodesPB := convert.PBFromInviteCode(list)

	res := &v1.GetInviteCodeResponse{
		Codes:      inviteCodesPB,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("GetInviteCodeLogic NewAny error: %v", err)
		return internal("系统繁忙，请稍后重试"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "获取成功",
		Data:    resAny,
	}, nil
}
