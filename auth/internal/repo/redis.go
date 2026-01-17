package repo

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type RedisRepo interface {
	SetEmailCode(email, code string, expire int) error
	GetEmailCode(email string) (string, error)
	DelEmailCode(email string) error
}

type RedisClient struct {
	client *redis.Redis
}

func NewRedisRepo(client *redis.Redis) RedisRepo {
	return &RedisClient{client: client}
}
