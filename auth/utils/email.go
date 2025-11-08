package utils

import (
	"fmt"
	"github.com/go-gomail/gomail"
)

const (
	EmailCodeLength = 6
)

// EmailConfig 邮箱配置
type EmailConfig struct {
	From     string
	Password string
	SMTPHost string
	SMTPPort int
}

// SendEmailCode 发送邮件验证码
func SendEmailCode(cfg EmailConfig, to, code string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", cfg.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "小安计划验证码")

	body := fmt.Sprintf(`
		<h2 style="color: #2E86C1;">小安计划验证码</h2>
		<p>你好！</p>
		<p>您的验证码是：</p>
		<p style="font-size: 24px; font-weight: bold; color: #E74C3C;">%s</p>
		<p>请在5分钟内使用。</p>
	`, code)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.From, cfg.Password)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}
	return nil
}

// GenerateEmailCode 生成随机验证码
func GenerateEmailCode() string {
	return GenerateCode(EmailCodeLength)
}
