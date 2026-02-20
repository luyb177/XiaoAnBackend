package like

import (
	"context"
	"errors"
	"fmt"

	redisV9 "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	TargetLikeKey = "like:target:%s:%d" // like:target:<content_type>:<content_id>
	UserLikeKey   = "like:user:%d"      // like:user:<user_id>
	MemberValue   = "%s:%d"             // <content_type>:<content_id>
)

type Repository interface {
	Like(ctx context.Context, userID uint64, contentType string, contentID uint64) (bool, error)
	Unlike(ctx context.Context, userID uint64, contentType string, contentID uint64) (bool, error)
	HasLiked(ctx context.Context, userID uint64, contentType string, contentID uint64) (bool, error)
	BatchHasLiked(ctx context.Context, userID uint64, contentType string, contentIDs []uint64) (map[uint64]bool, error)
}

type repo struct {
	rds *redis.Redis
}

func NewRepository(rds *redis.Redis) Repository {
	return &repo{
		rds: rds,
	}
}

// Like 点赞 第一个返回值表示 本次调用是否“真正产生了点赞行为” true 表示产生了点赞行为，false 表示用户之前已经点过赞了
func (r *repo) Like(ctx context.Context, userID uint64, contentType string, contentID uint64) (bool, error) {
	targetKey := targetLikeKey(contentType, contentID)
	userKey := userLikeKey(userID)
	member := memberValue(contentType, contentID)

	res, err := r.rds.EvalCtx(
		ctx,
		likeLua,
		[]string{userKey, targetKey},
		member,
		userID,
	)
	if err != nil {
		return false, err
	}

	code, ok := res.(int64)
	if !ok {
		return false, errors.New("invalid lua result")
	}

	// 已点过赞（幂等）
	if code == 0 {
		return false, nil
	}

	// 本次确实点赞成功
	return true, nil
}

// Unlike 取消点赞 第一个返回值表示 本次调用是否“真正产生了取消点赞行为” true 表示产生了取消点赞行为，false 表示用户之前并未点过赞
func (r *repo) Unlike(ctx context.Context, userID uint64, contentType string, contentID uint64) (bool, error) {
	// key
	targetKey := targetLikeKey(contentType, contentID)
	userKey := userLikeKey(userID)
	member := memberValue(contentType, contentID)

	res, err := r.rds.EvalCtx(
		ctx,
		unlikeLua,
		[]string{userKey, targetKey},
		member,
		userID,
	)
	if err != nil {
		return false, err
	}

	code, ok := res.(int64)
	if !ok {
		return false, errors.New("invalid lua result")
	}

	// 未点过赞，无法取消点赞（幂等）
	if code == 0 {
		return false, nil
	}

	return true, nil
}

func (r *repo) HasLiked(ctx context.Context, userID uint64, contentType string, contentID uint64) (bool, error) {
	// key
	userKey := userLikeKey(userID)
	member := memberValue(contentType, contentID)

	return r.rds.SismemberCtx(ctx, userKey, member)
}

func (r *repo) BatchHasLiked(ctx context.Context, userID uint64, contentType string, contentIDs []uint64) (map[uint64]bool, error) {
	results := make(map[uint64]bool, len(contentIDs))
	if len(contentIDs) == 0 {
		return results, nil
	}

	userKey := userLikeKey(userID)

	// 准备 members
	members := make([]interface{}, len(contentIDs))
	for i, id := range contentIDs {
		members[i] = memberValue(contentType, id)
	}

	var cmd *redisV9.BoolSliceCmd

	err := r.rds.PipelinedCtx(ctx, func(pipeliner redis.Pipeliner) error {
		cmd = pipeliner.SMIsMember(ctx, userKey, members...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 取结果
	boolRes, err := cmd.Result()
	if err != nil {
		return nil, err
	}
	for i, liked := range boolRes {
		results[contentIDs[i]] = liked
	}

	return results, nil

}

func targetLikeKey(contentType string, contentID uint64) string {
	return fmt.Sprintf(TargetLikeKey, contentType, contentID)
}

func userLikeKey(userID uint64) string {
	return fmt.Sprintf(UserLikeKey, userID)
}

func memberValue(contentType string, contentID uint64) string {
	return fmt.Sprintf(MemberValue, contentType, contentID)
}
