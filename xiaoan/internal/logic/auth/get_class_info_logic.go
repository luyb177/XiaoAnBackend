package auth

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	auth "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

type GetClassInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetClassInfoLogic 获取班级信息
func NewGetClassInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetClassInfoLogic {
	return &GetClassInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetClassInfoLogic) GetClassInfo(req *types.GetClassInfoRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.AuthRPC.GetClassInfo(l.ctx, &auth.GetClassInfoRequest{ClassId: req.ClassID})
	if err != nil {
		l.Errorf("rpc GetClassInfo error: %v", err)
		return logic.BadResponse("获取班级信息失败"), nil
	}

	var rpcData = &auth.GetClassInfoResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("rpc GetClassInfo UnmarshalTo error: %v", err)
		}
	}
	rpcClassInfo := rpcData.ClassInfo
	if rpcClassInfo == nil {
		rpcClassInfo = &auth.Class{}
	}

	httpClass := types.Class{
		ClassID:          rpcClassInfo.Id,
		ClassName:        rpcClassInfo.Name,
		ClassDescription: rpcClassInfo.Description,
		AdminID:          rpcClassInfo.AdminId,
		StudentCount:     rpcClassInfo.StudentCount,
		Status:           rpcClassInfo.Status,
		CreatedAt:        rpcClassInfo.CreatedAt,
		UpdatedAt:        rpcClassInfo.UpdatedAt,
	}

	httpData := types.GetClassInfoResponse{ClassInfo: httpClass}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
