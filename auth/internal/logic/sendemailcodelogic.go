package logic

import (
	"context"

	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	authcode "github.com/luyb177/XiaoAnBackend/auth/pkg/code"
	"github.com/luyb177/XiaoAnBackend/auth/pkg/email"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendEmailCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendEmailCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendEmailCodeLogic {
	return &SendEmailCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SendEmailCode 邮箱验证码
// 这里如果要返回 response 的话，是不能返回 error 的，所以要把err全换成nil
func (l *SendEmailCodeLogic) SendEmailCode(in *v1.SendEmailRequest) (*v1.Response, error) {
	if in.Email == "" {
		l.Errorf("SendEmailCode err: 邮箱不能为空")

		return &v1.Response{
			Code:    400,
			Message: "邮箱不能为空",
		}, nil
	}

	emailCfg := email.EmailConfig{
		From:     l.svcCtx.Config.Email.From,
		Password: l.svcCtx.Config.Email.Password,
		SMTPHost: l.svcCtx.Config.Email.SMTPHost,
		SMTPPort: l.svcCtx.Config.Email.SMTPPort,
	}

	code := authcode.EmailCode()

	go func() {
		err := l.svcCtx.RedisRepo.SetEmailCode(in.Email, code, 300)
		if err != nil {
			l.Errorf("设置邮件验证码失败: %v", err)
			return
		}

		if err = email.SendEmailCode(emailCfg, in.Email, code); err != nil {
			l.Errorf("发送邮件失败: %v", err)
			return
		}
	}()

	return &v1.Response{
		Code:    200,
		Message: "邮件发送中，请注意查收",
	}, nil
}
