package utils

const (
	InviteCodeLength = 8
)

func GenerateInviteCode() string {
	return GenerateCode(InviteCodeLength)
}
