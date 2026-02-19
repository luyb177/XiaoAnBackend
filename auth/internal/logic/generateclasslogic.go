package logic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	v1 "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"
)

type GenerateClassLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ClassDao model.ClassModel
}

func NewGenerateClassLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateClassLogic {
	return &GenerateClassLogic{
		ctx:      ctx,
		svcCtx:   svcCtx,
		Logger:   logx.WithContext(ctx),
		ClassDao: model.NewClassModel(svcCtx.Mysql),
	}
}

// GenerateClass 生成班级
func (l *GenerateClassLogic) GenerateClass(in *v1.GenerateClassRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || user.Role == "" || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或登录状态异常"), nil
	}

	if in.Name == "" || len(in.Name) > 50 {
		return bad("班级名称不能为空且不能超过50个字符"), nil
	}

	class := &model.Class{
		Name:         in.Name,
		AdminId:      user.UID,
		StudentCount: 0,
		Status:       ClassStatusNormal,
		DeletedAt:    0,
	}

	_, err := l.ClassDao.Insert(l.ctx, class)
	if err != nil {
		// 这里应该是不会有的，因为班级名称不需要唯一，但是如果有的话，说明数据库有问题，可能是唯一索引被误加了
		if errors.Is(err, model.ErrDuplicateEntry) {
			return bad("班级已存在"), nil
		}
		l.Errorf("GenerateClass Insert error: %v", err)
		return internal("班级创建失败"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "班级创建成功",
	}, nil
}
