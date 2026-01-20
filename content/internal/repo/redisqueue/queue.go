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
	payload, err := task.Payload()
	if err != nil {
		return err
	}
	rawTask := &RawTask{
		TaskID:    task.ID(),
		Retry:     0,
		MaxRetry:  MaxRetry,
		Data:      payload,
		CreatedAt: time.Now().Unix(),
	}

	rawTaskJson, err := json.Marshal(rawTask)
	if err != nil {
		return err
	}
	_, err = q.rds.LpushCtx(ctx, q.keys.Pending, string(rawTaskJson))
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
		// 这里 unmarshal 失败，则将任务放入 DLQ
		dlqTask := NewDLQTask(
			DLQReasonUnmarshal,
			DLQStageDequeue,
			err.Error(),
			data,
		)

		err = q.MoveToDLQWithReason(ctx, dlqTask)
		return nil, err
	}

	return rawTask, nil
}

// Ack 处理中队列 task -> 删除
func (q *RedisTaskQueue) Ack(ctx context.Context, task taskqueue.Task) error {
	payload, err := task.Payload()
	if err != nil {
		return err
	}
	_, err = q.rds.LremCtx(ctx, q.keys.Processing, 1, string(payload))
	return err
}

// Retry 处理中队列 task -> 延迟队列
func (q *RedisTaskQueue) Retry(ctx context.Context, task taskqueue.Task, delay time.Duration) error {
	payload, err := task.Payload()
	if err != nil {
		return err
	}

	var rawTask RawTask
	err = json.Unmarshal(payload, &rawTask)
	if err != nil {
		// 这里 unmarshal 失败，则将任务放入 DLQ
		dlqTask := NewDLQTask(
			DLQReasonUnmarshal,
			DLQStageRetry,
			err.Error(),
			payload,
		)
		return q.MoveToDLQWithReason(ctx, dlqTask)

	}

	if rawTask.Retry >= rawTask.MaxRetry {
		dlqTask := NewDLQTask(
			DLQReasonMaxRetry,
			DLQStageRetry,
			"retry count more than MaxRetry",
			payload,
		)
		return q.MoveToDLQWithReason(ctx, dlqTask)
	}

	rawTask.Retry++

	rawTask.DelaySec = int64(calculateDelay(delay, rawTask.Retry).Seconds())

	newPayload, err := json.Marshal(&rawTask)
	if err != nil {
		// 这里新的 payload marshal 失败，则将任务放入 DLQ
		dlqTask := NewDLQTask(
			DLQReasonMarshal,
			DLQStageRetry,
			err.Error(),
			payload,
		)
		return q.MoveToDLQWithReason(ctx, dlqTask)
	}

	score := time.Now().Unix() + rawTask.DelaySec

	_, err = q.rds.EvalCtx(
		ctx,
		moveProcessingToRetryLua,
		[]string{
			q.keys.Processing,
			q.keys.Retry,
		},
		string(payload),    // old payload
		string(newPayload), // new payload
		score,
	)
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

// MoveToDLQ  实现接口
func (q *RedisTaskQueue) MoveToDLQ(ctx context.Context, task taskqueue.Task) error {
	payload, err := task.Payload()
	if err != nil {
		return err
	}
	dlqTask := NewDLQTask(
		DLQReasonUnknown,
		DLQStageUnknown,
		"move to dlq by interface",
		payload,
	)
	return q.MoveToDLQWithReason(ctx, dlqTask)
}

func (q *RedisTaskQueue) MoveToDLQWithReason(ctx context.Context, dlqTask *DLQTask) error {
	b, marshalErr := json.Marshal(dlqTask)
	if marshalErr != nil {
		// 最终兜底：保证任务不会丢
		b = dlqTask.RawPayload
	}

	_, err := q.rds.EvalCtx(
		ctx,
		moveProcessingToDLQLua,
		[]string{
			q.keys.Processing,
			q.keys.DLQ,
		},
		string(dlqTask.RawPayload), // old payload
		string(b),                  // new payload
	)
	return err
}

// 最高 base 的 2^5 倍 的延迟
func calculateDelay(base time.Duration, retry int) time.Duration {
	return base * time.Duration(1<<(retry-1))
}
