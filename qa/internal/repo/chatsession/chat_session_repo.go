package chatsession

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	sessionKeyPattern = "chat:session:%d" // chat_session:{session_id} -> user_id
	sessionTTL        = 24 * 3600         // 1 day
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

type Repository interface {
	SetSessionOwner(ctx context.Context, sessionID, userID uint64) error
	GetSessionOwner(ctx context.Context, sessionID uint64) (uint64, error)
	DeleteSession(ctx context.Context, sessionID uint64) error
}

type repo struct {
	rds *redis.Redis
}

func NewRepository(rds *redis.Redis) Repository {
	return &repo{rds: rds}
}

func (r *repo) SetSessionOwner(ctx context.Context, sessionID, userID uint64) error {
	key := sessionKey(sessionID)
	return r.rds.SetexCtx(ctx, key, strconv.FormatUint(userID, 10), sessionTTL)
}
func (r *repo) GetSessionOwner(ctx context.Context, sessionID uint64) (uint64, error) {
	key := sessionKey(sessionID)

	val, err := r.rds.GetCtx(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, ErrSessionNotFound // 明确返回未命中
		}
		return 0, err
	}
	val = strings.TrimSpace(val)
	if val == "" {
		return 0, ErrSessionNotFound
	}

	uid, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf(
			"invalid uid in redis, key=%s val=%q: %w",
			key, val, err,
		)
	}

	return uid, nil
}

func (r *repo) DeleteSession(ctx context.Context, sessionID uint64) error {
	key := sessionKey(sessionID)
	_, err := r.rds.DelCtx(ctx, key)
	return err
}

func sessionKey(sessionID uint64) string {
	return fmt.Sprintf(sessionKeyPattern, sessionID)
}
