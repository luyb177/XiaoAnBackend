package worker

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/luyb177/XiaoAnBackend/infra/queue"
	"github.com/luyb177/XiaoAnBackend/infra/queue/redisqueue"
	"github.com/luyb177/XiaoAnBackend/qa/internal/model"
	"github.com/luyb177/XiaoAnBackend/qa/internal/svc"
	"github.com/luyb177/XiaoAnBackend/qa/pkg/taskqueue/tasks"
)

const (
	maxTitleLength = 256
)

type ChatSessionHandler struct {
	logx.Logger
	svcCtx         *svc.ServiceContext
	ChatSessionDao model.ChatSessionModel
}

func NewChatSessionHandler(ctx context.Context, svcCtx *svc.ServiceContext) *ChatSessionHandler {
	return &ChatSessionHandler{
		Logger:         logx.WithContext(ctx),
		svcCtx:         svcCtx,
		ChatSessionDao: model.NewChatSessionModel(svcCtx.Mysql),
	}
}

func (h *ChatSessionHandler) Handle(ctx context.Context, task queue.Task) error {
	payload, err := task.Payload()
	if err != nil {
		return err
	}

	var rawTask redisqueue.RawTask
	err = json.Unmarshal(payload, &rawTask)
	if err != nil {
		return err
	}

	chatSessionTask := tasks.ChatSessionTask{}
	err = json.Unmarshal(rawTask.Data, &chatSessionTask)
	if err != nil {
		return err
	}

	h.Infof("processing chat session task: %+v", chatSessionTask)

	switch chatSessionTask.Type {
	case tasks.ChatSessionUpdateTitle:
		return h.handleUpdateTitle(ctx, &chatSessionTask)
	default:
		h.Errorf("unknown task type: %+v", chatSessionTask.Type)
		return nil
	}
}

func (h *ChatSessionHandler) handleUpdateTitle(ctx context.Context, task *tasks.ChatSessionTask) error {
	userMessage := &openai.ChatCompletionUserMessageParam{
		Content: openai.ChatCompletionUserMessageParamContentUnion{
			OfString: param.NewOpt(task.UserMessage),
		},
	}
	title, err := h.svcCtx.LLMClient.ChatCompletionToTitle(ctx, userMessage)
	if err != nil {
		return err
	}
	if title == "" {
		return errors.New("empty title returned from LLM")
	}
	// 更新标题，按字符数截断以避免切到 UTF-8 中间字节
	if len([]rune(title)) > maxTitleLength {
		runes := []rune(title)
		title = string(runes[:maxTitleLength])
	}
	result, err := h.ChatSessionDao.UpdateTitle(ctx, task.SessionID, title)
	if err != nil {
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affect == 0 {
		// 这里没更新的话说明 session_id 不存在了
		h.Errorf("title %q for session %v not updated", title, task.SessionID)
	}
	return nil
}
