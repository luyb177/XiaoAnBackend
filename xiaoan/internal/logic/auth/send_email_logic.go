package auth

import (
	"context"
	"fmt"

	auth "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
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

// SendEmail 人为规定 res 不为 空
func (l *SendEmailLogic) SendEmail(req *types.SendEmailRequest) (resp *types.Response, err error) {
	if req.Email == "" {
		return &types.Response{
			Code:    400,
			Message: "邮箱不能为空",
		}, fmt.Errorf("邮箱不能为空")
	}

	res, _ := l.svcCtx.AuthRpc.SendEmailCode(l.ctx, &auth.SendEmailRequest{Email: req.Email})

	if res != nil {
		return &types.Response{
			Code:    res.Code,
			Message: res.Message,
		}, nil
	}

	// 兜底
	return &types.Response{
		Code:    400,
		Message: "发送失败",
	}, nil
}
