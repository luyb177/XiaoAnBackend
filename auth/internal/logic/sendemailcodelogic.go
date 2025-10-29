package logic

import (
	"context"
	"github.com/go-gomail/gomail"

	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pb/auth"

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
func (l *SendEmailCodeLogic) SendEmailCode(in *auth.SendEmailRequest) (*auth.Response, error) {
	sendEmail()
	return &auth.Response{}, nil
}

func sendEmail() {
	m := gomail.NewMessage()
	m.SetHeader("From", "3953017473@qq.com")
	m.SetHeader("To", "2085661244@qq.com")
	m.SetHeader("Subject", "测试邮件")
	m.SetBody("text/html", "<b>你好，这是测试邮件！</b>")

	d := gomail.NewDialer("smtp.example.com", 587, "3953017473@qq.com", "towxkhqxeqnlccib")

	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}
