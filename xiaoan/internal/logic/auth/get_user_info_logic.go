package auth

import (
	"context"

	auth "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetUserInfoLogic 获取用户信息
func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserInfoLogic) GetUserInfo(req *types.GetUserInfoRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.AuthRpc.GetUserInfo(l.ctx, &auth.GetUserInfoRequest{UserId: req.UserID})
	if err != nil {
		l.Errorf("rpc GetUserInfo err: %v", err)
		return logic.BadResponse("获取用户信息失败"), nil
	}

	var rpcData = &auth.GetUserInfoResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("rpc GetUserInfo UnmarshalTo err: %v", err)
		}
	}
	rpcUserInfo := rpcData.User
	if rpcUserInfo == nil {
		rpcUserInfo = &auth.UserInfo{}
	}

	httpUserInfo := types.UserInfo{
		UserID:     rpcUserInfo.Id,
		Name:       rpcUserInfo.Name,
		Email:      rpcUserInfo.Email,
		Avatar:     rpcUserInfo.Avatar,
		Phone:      rpcUserInfo.Phone,
		Department: rpcUserInfo.Department,
		Role:       rpcUserInfo.Role,
		ClassID:    rpcUserInfo.ClassId,
		Status:     rpcUserInfo.Status,
		CreatedAt:  rpcUserInfo.CreatedAt,
		UpdatedAt:  rpcUserInfo.UpdatedAt,
	}

	httpData := types.GetUserInfoResponse{UserInfo: httpUserInfo}

	return &types.Response{
		Code:    200,
		Message: "获取用户信息成功",
		Data:    httpData,
	}, nil
}
