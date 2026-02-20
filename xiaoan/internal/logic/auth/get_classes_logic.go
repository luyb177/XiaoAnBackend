package auth

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	auth "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

type GetClassesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetClassesLogic 获取班级列表
func NewGetClassesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetClassesLogic {
	return &GetClassesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetClassesLogic) GetClasses(req *types.GetClassesRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.AuthRPC.GetClasses(l.ctx, &auth.GetClassesRequest{
		PageSize: req.PageSize,
		Cursor:   req.Cursor,
		UserId:   req.UserID,
	})

	if err != nil {
		l.Errorf("rpc GetClasses error: %v", err)
		return logic.BadResponse("获取班级列表失败"), nil
	}

	var rpcData = &auth.GetClassesResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("rpc GetClasses UnmarshalTo error: %v", err)
		}
	}

	rpcClasses := rpcData.Classes
	if rpcClasses == nil {
		rpcClasses = []*auth.Class{}
	}

	httpClasses := make([]types.Class, len(rpcClasses))
	for i, rpcClass := range rpcClasses {
		if rpcClass == nil {
			rpcClass = &auth.Class{}
		}

		httpClasses[i] = types.Class{
			ClassID:          rpcClass.Id,
			ClassName:        rpcClass.Name,
			ClassDescription: rpcClass.Description,
			AdminID:          rpcClass.AdminId,
			StudentCount:     rpcClass.StudentCount,
			Status:           rpcClass.Status,
			CreatedAt:        rpcClass.CreatedAt,
			UpdatedAt:        rpcClass.UpdatedAt,
		}
	}

	httpData := types.GetClassesResponse{
		Classes:    httpClasses,
		HasMore:    rpcData.HasMore,
		NextCursor: rpcData.NextCursor,
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
