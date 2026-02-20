package logic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/auth/pkg/user/convert"
	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"
)

type GetClassMembersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	UserDao model.UserModel
}

func NewGetClassMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetClassMembersLogic {
	return &GetClassMembersLogic{
		ctx:     ctx,
		svcCtx:  svcCtx,
		Logger:  logx.WithContext(ctx),
		UserDao: model.NewUserModel(svcCtx.Mysql),
	}
}

// GetClassMembers 获取班级成员列表
func (l *GetClassMembersLogic) GetClassMembers(in *v1.GetClassMembersRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || user.Role == "" || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或登录状态异常"), nil
	}

	if in.ClassId == constants.InvalidClassID {
		return bad("参数错误"), nil
	}

	if in.PageSize <= 0 {
		in.PageSize = 10
	}

	limit := in.PageSize + 1

	var (
		list []*model.User
		err  error
	)

	if in.Cursor == 0 {
		// 首次查询
		list, err = l.UserDao.FindManyByClassID(l.ctx, in.ClassId, limit)
	} else {
		// 翻页查询，cursor为上次查询结果的最后一个用户ID
		list, err = l.UserDao.FindManyByClassIDWithCursor(l.ctx, in.ClassId, in.Cursor, limit)
	}

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return notFound("班级不存在或没有成员"), nil
		}
		l.Errorf("查询班级成员列表失败: %v", err)
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

	userInfosPB := convert.PBFromUsers(list)

	res := &v1.GetClassMembersResponse{
		Members:    userInfosPB,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("anypb.New GetClassMembersResponse error: %v", err)
		return internal("系统繁忙，请稍后再试"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "获取班级成员列表成功",
		Data:    resAny,
	}, nil
}
