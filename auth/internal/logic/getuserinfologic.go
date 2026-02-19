package logic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	v1 "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"
)

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	UserDao model.UserModel
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:     ctx,
		svcCtx:  svcCtx,
		Logger:  logx.WithContext(ctx),
		UserDao: model.NewUserModel(svcCtx.Mysql),
	}
}

// GetUserInfo 获取用户信息
func (l *GetUserInfoLogic) GetUserInfo(in *v1.GetUserInfoRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || user.Role == "" || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或登录状态异常"), nil
	}

	targetUser, err := l.UserDao.FindOneWithNotDelete(l.ctx, in.UserId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return notFound("用户不存在"), nil
		}
		l.Errorf("GetUserInfoLogic FindOneWithNotDelete err: %v", err)
		return internal("系统繁忙，请稍后再试"), nil
	}

	res := &v1.GetUserInfoResponse{User: &v1.UserInfo{
		Id:         targetUser.Id,
		Name:       targetUser.Name,
		Email:      targetUser.Email,
		Avatar:     targetUser.Avatar.String,
		Phone:      targetUser.Phone.String,
		Department: targetUser.Department.String,
		Role:       targetUser.Role,
		ClassId:    targetUser.ClassId,
		Status:     targetUser.Status,
		CreatedAt:  targetUser.CreatedAt.Unix(),
		UpdatedAt:  targetUser.UpdatedAt.Unix(),
	}}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("GetUserInfo anypb.New err: %v", err)
		return internal("系统繁忙，请稍后再试"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "获取用户基本信息成功",
		Data:    resAny,
	}, nil
}
