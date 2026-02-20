package auth

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	auth "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

type GenerateClassLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 生成班级
func NewGenerateClassLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateClassLogic {
	return &GenerateClassLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateClassLogic) GenerateClass(req *types.GenerateClassRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.AuthRPC.GenerateClass(
		l.ctx,
		&auth.GenerateClassRequest{
			Name:        req.ClassName,
			Description: req.ClassDescription,
		})

	if err != nil {
		l.Errorf("rpc GenerateClass error: %v", err)
		return logic.BadResponse("班级创建失败"), nil
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    &types.EmptyResponse{},
	}, nil
}
