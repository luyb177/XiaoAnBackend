package worker

import (
	"context"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/luyb177/XiaoAnBackend/infra/queue"
	"github.com/luyb177/XiaoAnBackend/qa/internal/svc"
	"github.com/luyb177/XiaoAnBackend/qa/pkg/taskqueue/tasks"
)

type Worker struct {
	logx.Logger
	taskQueue queue.TaskQueue
	handlers  map[tasks.TaskPrefix]TaskHandler

	// 用于实现 service 的接口
	ctx    context.Context
	cancel context.CancelFunc
}

// Start implements service.Service
func (w *Worker) Start() {
	w.Info("worker started")

	if err := w.start(w.ctx); err != nil && !errors.Is(err, context.Canceled) {
		w.Errorf("worker exited with error: %v", err)
	}
}

// Stop implements service.Service
func (w *Worker) Stop() {
	if w.cancel != nil {
		w.Info("worker stopping")
		w.cancel()
	}
}

func NewWorker(svcCtx *svc.ServiceContext) *Worker {
	w := &Worker{
		taskQueue: svcCtx.TaskQueue,
		handlers:  make(map[tasks.TaskPrefix]TaskHandler),
		Logger:    logx.WithContext(context.Background()),
	}

	w.ctx, w.cancel = context.WithCancel(context.Background())

	// 注册处理器
	w.RegisterHandler(tasks.ChatSessionTaskPrefix, NewChatSessionHandler(w.ctx, svcCtx))

	return w
}

type TaskHandler interface {
	Handle(ctx context.Context, task queue.Task) error
}

func (w *Worker) RegisterHandler(taskType tasks.TaskPrefix, handler TaskHandler) {
	w.handlers[taskType] = handler
}

// start 启动 Worker
func (w *Worker) start(ctx context.Context) error {
	// 启动重试任务迁移的定时任务
	go w.startRetryScheduler(ctx)

	// 启动任务消费循环
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := w.processTask(ctx); err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				w.Errorf("process task error: %v", err)
			}
		}
	}
}

func (w *Worker) processTask(ctx context.Context) error {
	task, err := w.taskQueue.Dequeue(ctx)
	if err != nil {
		return err
	}

	if task == nil {
		return nil // 队列为空
	}

	// 根据任务 ID 前缀找到对应的处理器
	handler := w.findHandler(task.ID())
	if handler == nil {
		w.Errorf("no handler found for task: %s", task.ID())
		return w.taskQueue.MoveToDLQ(ctx, task)
	}

	// 执行任务
	if err := handler.Handle(ctx, task); err != nil {
		w.Errorf("handle task error: %s", err.Error())
		// 重试
		return w.taskQueue.Retry(ctx, task, 10*time.Second)
	}

	// 确认任务完成
	return w.taskQueue.Ack(ctx, task)
}

func (w *Worker) findHandler(taskID string) TaskHandler {
	// 前缀匹配
	for prefix, handler := range w.handlers {
		if len(taskID) >= len(prefix) && tasks.TaskPrefix(taskID[:len(prefix)]) == prefix {
			return handler
		}
	}
	return nil
}

// startRetryScheduler 定时将延迟队列中到期的任务移回待处理队列
func (w *Worker) startRetryScheduler(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.taskQueue.MoveRetryToPending(ctx); err != nil {
				w.Errorf("move retry to pending error: %v", err)
			}
		}
	}
}
