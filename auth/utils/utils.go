package utils

import (
	"crypto/rand"
	"math/big"
)

const (
	letters = "23456789abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"
)

func GenerateCode(length int) string {
	code := make([]byte, length)
	for i := range code {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		code[i] = letters[num.Int64()]
	}
	return string(code)
}
