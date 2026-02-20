package auth

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	auth "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

type SendEmailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendEmailLogic {
	return &SendEmailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// SendEmail 人为规定 res 不为 空
func (l *SendEmailLogic) SendEmail(req *types.SendEmailRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.AuthRPC.SendEmailCode(l.ctx, &auth.SendEmailRequest{Email: req.Email})
	if err != nil {
		l.Errorf("rpc SendEmailCode err: %s", err.Error())
		return logic.BadResponse("发送失败"), nil
	}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    &types.EmptyResponse{},
	}, nil
}
