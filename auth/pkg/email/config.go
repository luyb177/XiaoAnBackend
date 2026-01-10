package email

// EmailConfig 邮箱配置
type EmailConfig struct {
	From     string
	Password string
	SMTPHost string
	SMTPPort int
}
