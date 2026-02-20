package logic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/auth/pkg/class/convert"
	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"
)

type GetClassesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ClassDao model.ClassModel
}

func NewGetClassesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetClassesLogic {
	return &GetClassesLogic{
		ctx:      ctx,
		svcCtx:   svcCtx,
		Logger:   logx.WithContext(ctx),
		ClassDao: model.NewClassModel(svcCtx.Mysql),
	}
}

// GetClasses 获取班级列表
func (l *GetClassesLogic) GetClasses(in *v1.GetClassesRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || user.Role == "" || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或登录状态异常"), nil
	}

	if in.PageSize <= 0 {
		in.PageSize = 10
	}

	var targetUserID uint64
	if user.UID != in.UserId && in.UserId != 0 {
		if user.Role != constants.SUPERADMIN && user.Role != constants.STAFF {
			return bad("没有权限查询其他用户创建的班级"), nil
		}
		targetUserID = in.UserId
	} else {
		targetUserID = user.UID
	}

	// 查询多一条记录，来判断是否有下一页
	limit := in.PageSize + 1

	var (
		list []*model.Class
		err  error
	)

	if in.Cursor == 0 {
		// 首次查询
		list, err = l.ClassDao.FindManyByUserID(l.ctx, targetUserID, limit)
	} else {
		// 继续查询
		list, err = l.ClassDao.FindManyByUserIDWithCursor(l.ctx, targetUserID, in.Cursor, limit)
	}
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return notFound("没有找到相应的班级"), nil
		}
		l.Errorf("GetClassesLogic FindManyByUserIDWithCursor error: %v", err)
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

	classesPB := convert.PBFromClass(list)

	res := &v1.GetClassesResponse{
		Classes:    classesPB,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("GetClassesLogic NewAny error: %v", err)
		return internal("系统繁忙，请稍后重试"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "查询成功",
		Data:    resAny,
	}, nil
}
