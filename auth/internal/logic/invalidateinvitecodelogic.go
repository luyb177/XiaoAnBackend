package logic

import (
	"context"
	"errors"
	"strings"

	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"

	"github.com/zeromicro/go-zero/core/logx"
)

type InvalidateInviteCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	InviteCodeDao model.InviteCodeModel
}

func NewInvalidateInviteCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InvalidateInviteCodeLogic {
	return &InvalidateInviteCodeLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		InviteCodeDao: model.NewInviteCodeModel(svcCtx.Mysql),
	}
}

// InvalidateInviteCode 失效邀请码
func (l *InvalidateInviteCodeLogic) InvalidateInviteCode(in *v1.InvalidateInviteCodeRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || user.Role == "" || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或登录状态异常"), nil
	}

	if strings.TrimSpace(in.Code) == "" {
		return bad("邀请码不能为空"), nil
	}

	// 查询一下邀请码信息
	inviteCode, err := l.InviteCodeDao.FindUsableByCode(l.ctx, in.Code)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return notFound("邀请码不存在或已失效"), nil
		}
		l.Errorf("FindUsableByCode error: %v", err)
		return internal("系统繁忙，请稍后再试"), nil
	}

	if inviteCode.CreatorId != user.UID {
		if user.Role != constants.SUPERADMIN && user.Role != constants.STAFF {
			return bad("没有权限失效其他用户创建的邀请码"), nil
		}
	}

	result, err := l.InviteCodeDao.InvalidateInviteCode(l.ctx, inviteCode.Id)
	if err != nil {
		l.Errorf("InvalidateInviteCode error: %v", err)
		return internal("系统繁忙，请稍后再试"), nil
	}
	affect, err := result.RowsAffected()
	if err != nil {
		l.Errorf("RowsAffected error: %v", err)
		return internal("系统繁忙，请稍后再试"), nil
	}
	if affect == 0 {
		return notFound("邀请码不存在或已失效"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "邀请码已失效",
	}, nil
}
