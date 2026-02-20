package email

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	CodeKey = "email:code:%s"
)

type Repository interface {
	SetEmailCode(emailCanonical, code string, expire time.Duration) error
	GetEmailCode(emailCanonical string) (string, bool, error)
	DelEmailCode(emailCanonical string) error
}

type repo struct {
	client *redis.Redis
}

func NewRedisRepo(client *redis.Redis) Repository {
	return &repo{client: client}
}

func (r *repo) SetEmailCode(emailCanonical, code string, expire time.Duration) error {
	key := emailCodeKey(emailCanonical)

	return r.client.Setex(key, code, int(expire.Seconds()))
}

func (r *repo) GetEmailCode(emailCanonical string) (string, bool, error) {
	key := emailCodeKey(emailCanonical)

	val, err := r.client.Get(key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", false, nil
		}
		return "", false, err
	}
	return val, true, nil
}

func (r *repo) DelEmailCode(emailCanonical string) error {
	key := emailCodeKey(emailCanonical)

	_, err := r.client.Del(key)
	return err
}

func emailCodeKey(emailCanonical string) string {
	sum := sha256.Sum256([]byte(emailCanonical))
	return fmt.Sprintf(CodeKey, hex.EncodeToString(sum[:]))
}
