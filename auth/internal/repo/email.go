package repo

import (
	"fmt"
)

const (
	EmailCodeKey = "email:code:%s"
)

func (r *RedisClient) SetEmailCode(email, code string, expire int) error {
	key := fmt.Sprintf(EmailCodeKey, email)
	return r.client.Setex(key, code, expire)
}

func (r *RedisClient) GetEmailCode(email string) (string, error) {
	key := fmt.Sprintf(EmailCodeKey, email)
	return r.client.Get(key)
}

func (r *RedisClient) DelEmailCode(email string) error {
	key := fmt.Sprintf(EmailCodeKey, email)
	_, err := r.client.Del(key)
	return err
}
