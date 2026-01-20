package worker

import (
	"context"
	"time"

	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"

	"github.com/zeromicro/go-zero/core/logx"
)

type Worker struct {
	logx.Logger
	taskQueue taskqueue.TaskQueue
	handlers  map[tasks.TaskPrefix]TaskHandler
}

type TaskHandler interface {
	Handle(ctx context.Context, task taskqueue.Task) error
}

func NewWorker(svcCtx *svc.ServiceContext) *Worker {
	w := &Worker{
		taskQueue: svcCtx.TaskQueue,
		handlers:  make(map[tasks.TaskPrefix]TaskHandler),
	}

	// 注册处理器
	w.RegisterHandler(tasks.ArticleRelationTaskPrefix, NewArticleRelationHandler(svcCtx))

	return w
}

func (w *Worker) RegisterHandler(taskType tasks.TaskPrefix, handler TaskHandler) {
	w.handlers[taskType] = handler
}

// Start 启动 Worker
func (w *Worker) Start(ctx context.Context) error {
	// 启动重试任务迁移的定时任务
	go w.startRetryScheduler(ctx)

	// 启动任务消费循环
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := w.processTask(ctx); err != nil {
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
