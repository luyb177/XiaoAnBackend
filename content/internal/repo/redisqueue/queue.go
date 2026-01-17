package redisqueue

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	RIGHT = "RIGHT"
	LEFT  = "LEFT"
)

const (
	MaxRetry = 5
)

// RawTask 通用 task 的实现
type RawTask struct {
	TaskID    string          `json:"task_id"`    // 幂等 ID
	Retry     int             `json:"retry"`      // 当前重试次数
	MaxRetry  int             `json:"max_retry"`  // 最大重试次数
	DelaySec  int64           `json:"delay_sec"`  // 下一次 retry 的延迟（秒）
	Data      json.RawMessage `json:"data"`       // 业务任务
	CreatedAt int64           `json:"created_at"` // 创建时间
}

func (t *RawTask) ID() string {
	return t.TaskID
}

func (t *RawTask) Payload() []byte {
	payload, _ := json.Marshal(t)
	return payload
}

type RedisTaskQueue struct {
	rds          *redis.Redis
	keys         taskqueue.QueueKey
	blockingNode redis.RedisNode
}

func NewRedisTaskQueue(rds *redis.Redis, keys taskqueue.QueueKey) *RedisTaskQueue {
	node, err := redis.CreateBlockingNode(rds)
	if err != nil {
		panic(err)
	}

	return &RedisTaskQueue{
		rds:          rds,
		keys:         keys,
		blockingNode: node,
	}
}

// Enqueue  task -> 待处理队列
func (q *RedisTaskQueue) Enqueue(ctx context.Context, task taskqueue.Task) error {
	rawTask := &RawTask{
		TaskID:    task.ID(),
		Retry:     0,
		MaxRetry:  MaxRetry,
		Data:      task.Payload(),
		CreatedAt: time.Now().Unix(),
	}

	rawTaskJson, _ := json.Marshal(rawTask)
	_, err := q.rds.LpushCtx(ctx, q.keys.Pending, string(rawTaskJson))
	return err
}

// Dequeue 待处理队列 -> task -> 处理中队列
func (q *RedisTaskQueue) Dequeue(ctx context.Context) (taskqueue.Task, error) {
	// 堵塞 原子性
	data, err := q.blockingNode.BLMove(
		ctx,
		q.keys.Pending,    // source
		q.keys.Processing, // destination
		RIGHT,             // source direction 原始位置
		LEFT,              // destination direction 目标位置
		5*time.Second,
	).Bytes()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	rawTask := &RawTask{}
	err = json.Unmarshal(data, rawTask)
	if err != nil {
		return nil, err
	}

	return &RawTask{
		TaskID:    rawTask.TaskID,
		Retry:     rawTask.Retry,
		MaxRetry:  rawTask.MaxRetry,
		DelaySec:  rawTask.DelaySec,
		Data:      rawTask.Data,
		CreatedAt: rawTask.CreatedAt,
	}, nil
}

// Ack 处理中队列 task -> 删除
func (q *RedisTaskQueue) Ack(ctx context.Context, task taskqueue.Task) error {
	_, err := q.rds.LremCtx(ctx, q.keys.Processing, 1, string(task.Payload()))
	return err
}

// Retry 处理中队列 task -> 延迟队列
func (q *RedisTaskQueue) Retry(ctx context.Context, task taskqueue.Task, delay time.Duration) error {
	var rawTask RawTask
	err := json.Unmarshal(task.Payload(), &rawTask)
	if err != nil {
		return err
	}
	rawTask.Retry += 1
	if rawTask.Retry > rawTask.MaxRetry {
		return q.MoveToDLQ(ctx, task)
	}

	score := time.Now().Add(delay).Unix()
	_, err = q.rds.ZaddCtx(ctx, q.keys.Retry, score, string(task.Payload()))
	return err
}

// MoveRetryToPending 延迟队列 task -> 待处理队列
func (q *RedisTaskQueue) MoveRetryToPending(ctx context.Context) error {
	now := time.Now().Unix()

	_, err := q.rds.EvalCtx(
		ctx,
		moveRetryToPendingLua,
		[]string{
			q.keys.Retry,
			q.keys.Pending,
		},
		now,
	)
	return err
}

func (q *RedisTaskQueue) MoveToDLQ(ctx context.Context, task taskqueue.Task) error {
	_, err := q.rds.LpushCtx(ctx, q.keys.DLQ, string(task.Payload()))
	return err
}
