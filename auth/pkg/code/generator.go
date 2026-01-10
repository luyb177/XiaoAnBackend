package code

import (
	"crypto/rand"
	"math/big"
)

func Generate(length int) string {
	code := make([]byte, length)
	for i := range code {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		code[i] = letters[num.Int64()]
	}
	return string(code)
}

// EmailCode 生成随机验证码
func EmailCode() string {
	return Generate(EmailCodeLength)
}

// InviteCode 邀请码
func InviteCode() string {
	return Generate(InviteCodeLength)
}
