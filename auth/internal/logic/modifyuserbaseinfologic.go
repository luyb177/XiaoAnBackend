package logic

import (
	"context"
	"database/sql"
	"errors"

	"github.com/luyb177/XiaoAnBackend/auth/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyUserBaseInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	UserDao model.UserModel
}

func NewModifyUserBaseInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyUserBaseInfoLogic {
	return &ModifyUserBaseInfoLogic{
		ctx:     ctx,
		svcCtx:  svcCtx,
		Logger:  logx.WithContext(ctx),
		UserDao: model.NewUserModel(svcCtx.Mysql),
	}
}

// ModifyUserBaseInfo 修改用户基本信息
func (l *ModifyUserBaseInfoLogic) ModifyUserBaseInfo(in *v1.ModifyUserBaseInfoRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == InvalidUserID || user.Role == "" || user.Status != UserStatusNormal {
		return bad("用户未登录或登录状态异常"), nil
	}

	if resp := l.validate(in); resp != nil {
		return resp, nil
	}

	// 查询用户
	targetUser, err := l.UserDao.FindOneWithNotDelete(l.ctx, in.UserId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return notFound("用户不存在"), nil
		}
		l.Errorf("FindOneWithNotDelete error: %v", err)
		return internal("用户信息查询失败"), nil
	}

	if user.UID != targetUser.Id {
		if user.Role == STUDENT {
			return bad("无权限修改该用户信息"), nil
		}
	}

	if in.Name != "" {
		targetUser.Name = in.Name
	}
	if in.Phone != "" {
		targetUser.Phone = sql.NullString{String: in.Phone, Valid: true}
	}
	if in.Avatar != "" {
		targetUser.Avatar = sql.NullString{String: in.Avatar, Valid: true}
	}

	err = l.UserDao.Update(l.ctx, targetUser)
	if err != nil {
		l.Errorf("Update error: %v", err)
		return internal("用户信息更新失败"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "用户信息更新成功",
	}, nil
}

func (l *ModifyUserBaseInfoLogic) validate(in *v1.ModifyUserBaseInfoRequest) *v1.Response {
	if in.UserId == InvalidUserID {
		return bad("用户ID不合法")
	}
	if in.Name != "" && len(in.Name) > 50 {
		return bad("用户名长度不能超过50个字符")
	}
	return nil
}
