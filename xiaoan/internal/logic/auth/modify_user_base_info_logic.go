package auth

import (
	"context"
	auth "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyUserBaseInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewModifyUserBaseInfoLogic 修改用户基本信息
func NewModifyUserBaseInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyUserBaseInfoLogic {
	return &ModifyUserBaseInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ModifyUserBaseInfoLogic) ModifyUserBaseInfo(req *types.ModifyUserBaseInfoRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.AuthRpc.ModifyUserBaseInfo(l.ctx, &auth.ModifyUserBaseInfoRequest{
		Name:   req.Name,
		Avatar: req.Avatar,
		Phone:  req.Phone,
		UserId: req.UserID,
	})

	if err != nil {
		l.Errorf("rpc ModifyUserBaseInfo err: %v", err)
		return logic.BadResponse("修改用户基本信息失败"), nil
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    &types.EmptyResponse{},
	}, nil
}
