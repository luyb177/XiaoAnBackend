package chatmessage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	messageListKeyPattern = "chat:msg:list:%d" // chat:msg:list:{session_id}
	MaxMessageListSize    = 30
	messageListTTL        = 24 * 3600 // 1 day
)

type Repository interface {
	UpsertMessage(ctx context.Context, sessionID uint64, msg *MessageCache) error
	GetRecentMessages(ctx context.Context, sessionID uint64) ([]*MessageCache, error)
}

type repo struct {
	rds *redis.Redis
}

func NewRepository(rds *redis.Redis) Repository {
	return &repo{
		rds: rds,
	}
}

type MessageCache struct {
	ID        uint64 `json:"id"`
	MessageID string `json:"message_id"`
	Role      int64  `json:"role"`
	Content   string `json:"content"`
	Ts        int64  `json:"ts"`
}

func (r *repo) UpsertMessage(ctx context.Context, sessionID uint64, msg *MessageCache) error {
	key := targetMessageListKey(sessionID)

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = r.rds.EvalCtx(ctx,
		upsertMessageLua,
		[]string{key},
		data,
		MaxMessageListSize,
		messageListTTL,
	)
	return err
}

func (r *repo) GetRecentMessages(ctx context.Context, sessionID uint64) ([]*MessageCache, error) {
	key := targetMessageListKey(sessionID)

	vals, err := r.rds.LrangeCtx(ctx, key, 0, MaxMessageListSize-1)
	if err != nil {
		return nil, err
	}

	res := make([]*MessageCache, 0, len(vals))
	// 倒序解析，保证返回的消息列表是从旧到新的顺序
	for i := len(vals) - 1; i >= 0; i-- {
		var m MessageCache
		if err = json.Unmarshal([]byte(vals[i]), &m); err != nil {
			logx.Errorf("unmarshal message cache failed: %v", err)
			continue
		}
		res = append(res, &m)
	}
	return res, nil
}

func targetMessageListKey(sessionID uint64) string {
	return fmt.Sprintf(messageListKeyPattern, sessionID)
}
