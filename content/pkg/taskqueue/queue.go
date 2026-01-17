package taskqueue

import (
	"context"
	"time"
)

type Task interface {
	ID() string      // 唯一标识
	Payload() []byte // 任务序列化内容
}

// QueueKey 统一管理 队列 的 key
type QueueKey struct {
	Pending    string // pending - 待处理队列
	Processing string // processing - 处理中队列
	Retry      string // retry - 重试队列
	DLQ        string // dlq - 死信队列
}

type TaskQueue interface {
	Enqueue(ctx context.Context, task Task) error
	Dequeue(ctx context.Context) (Task, error)
	Ack(ctx context.Context, task Task) error
	Retry(ctx context.Context, task Task, delay time.Duration) error
	MoveRetryToPending(ctx context.Context) error
	MoveToDLQ(ctx context.Context, task Task) error
}
