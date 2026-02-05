package collect

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	TargetCollectKey = "collect:target:%s:%d" // collect:target:<content_type>:<content_id>
	UserCollectKey   = "collect:user:%d"      // collect:user:<user_id>
	MemberValue      = "%s:%d"                // <content_type>:<content_id>
)

type Repository interface {
	Collect(ctx context.Context, userID uint64, contentType string, contentID uint64) (bool, error)
	UnCollect(ctx context.Context, userID uint64, contentType string, contentID uint64) (bool, error)
	HasCollect(ctx context.Context, userID uint64, contentType string, contentID uint64) (bool, error)
}

type repo struct {
	rds *redis.Redis
}

func NewRepository(rds *redis.Redis) Repository {
	return &repo{
		rds: rds,
	}
}

func (r *repo) Collect(ctx context.Context, userID uint64, contentType string, contentID uint64) (bool, error) {
	targetKey := targetCollectKey(contentType, contentID)
	userKey := userCollectKey(userID)
	member := memberValue(contentType, contentID)

	res, err := r.rds.EvalCtx(
		ctx,
		collectLua,
		[]string{userKey, targetKey},
		member,
		userID,
	)
	if err != nil {
		return false, err
	}

	code, ok := res.(int64)
	if !ok {
		return false, fmt.Errorf("invalid lua result")
	}

	// 已收藏（幂等）
	if code == 0 {
		return false, nil
	}

	// 本次确实收藏成功
	return true, nil
}

func (r *repo) UnCollect(ctx context.Context, userID uint64, contentType string, contentID uint64) (bool, error) {
	targetKey := targetCollectKey(contentType, contentID)
	userKey := userCollectKey(userID)
	member := memberValue(contentType, contentID)

	res, err := r.rds.EvalCtx(
		ctx,
		unCollectLua,
		[]string{userKey, targetKey},
		member,
		userID,
	)
	if err != nil {
		return false, err
	}

	code, ok := res.(int64)
	if !ok {
		return false, fmt.Errorf("invalid lua result")
	}

	// 未收藏（幂等）
	if code == 0 {
		return false, nil
	}

	// 本次确实取消收藏成功
	return true, nil
}

func (r *repo) HasCollect(ctx context.Context, userID uint64, contentType string, contentID uint64) (bool, error) {
	userKey := userCollectKey(userID)
	member := memberValue(contentType, contentID)

	return r.rds.SismemberCtx(ctx, userKey, member)
}

func targetCollectKey(contentType string, contentID uint64) string {
	return fmt.Sprintf(TargetCollectKey, contentType, contentID)
}

func userCollectKey(userID uint64) string {
	return fmt.Sprintf(UserCollectKey, userID)
}

func memberValue(contentType string, contentID uint64) string {
	return fmt.Sprintf(MemberValue, contentType, contentID)
}
