package auth

import (
	"context"

	pb "github.com/luyb177/XiaoAnBackend/auth/pb/auth"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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

func (l *SendEmailLogic) SendEmail(req *types.SendEmailRequest) (resp *types.Response, err error) {
	if req.Email == "" {
		return &types.Response{
			Code:    400,
			Message: "邮箱不能为空",
		}, nil
	}
	res, err := l.svcCtx.AuthRpc.SendEmailCode(l.ctx, &pb.SendEmailRequest{Email: req.Email})
	if err != nil {
		return &types.Response{
			Code:    400,
			Message: "发送失败",
		}, nil
	}
	return &types.Response{
		Code:    res.Code,
		Message: res.Message,
		Data:    res.Data,
	}, nil
}
