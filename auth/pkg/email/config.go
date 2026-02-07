package email

// Config 邮箱配置
type Config struct {
	From     string
	Password string
	SMTPHost string
	SMTPPort int
}
