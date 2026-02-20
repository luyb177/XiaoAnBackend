package logic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"
)

type GetClassInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ClassDao model.ClassModel
}

func NewGetClassInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetClassInfoLogic {
	return &GetClassInfoLogic{
		ctx:      ctx,
		svcCtx:   svcCtx,
		Logger:   logx.WithContext(ctx),
		ClassDao: model.NewClassModel(svcCtx.Mysql),
	}
}

// GetClassInfo 获取班级信息
func (l *GetClassInfoLogic) GetClassInfo(in *v1.GetClassInfoRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || user.Role == "" || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或登录状态异常"), nil
	}

	if in.ClassId == constants.InvalidClassID {
		return bad("参数错误"), nil
	}

	class, err := l.ClassDao.FindNormalOne(l.ctx, in.ClassId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return notFound("班级不存在"), nil
		}
		return internal("系统繁忙，请稍后再试"), nil
	}

	res := &v1.GetClassInfoResponse{ClassInfo: &v1.Class{
		Id:           class.Id,
		Name:         class.Name,
		Description:  class.Description.String,
		AdminId:      class.AdminId,
		StudentCount: class.StudentCount,
		Status:       class.Status,
		CreatedAt:    class.CreatedAt.Unix(),
		UpdatedAt:    class.UpdatedAt.Unix(),
	}}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("anypb.New failed: %v", err)
		return internal("系统繁忙，请稍后再试"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "success",
		Data:    resAny,
	}, nil
}
