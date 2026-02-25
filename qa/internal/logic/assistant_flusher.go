package logic

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/luyb177/XiaoAnBackend/qa/internal/model"
	"github.com/zeromicro/go-zero/core/logx"
)

type assistantFlusher struct {
	log logx.Logger
	dao model.ChatMessageModel

	msg *model.ChatMessage

	temp        strings.Builder
	lastFlushAt time.Time
	flushEvery  time.Duration
	flushChars  int
}

func newAssistantFlusher(log logx.Logger, dao model.ChatMessageModel, msg *model.ChatMessage) *assistantFlusher {
	return &assistantFlusher{
		log:         log,
		dao:         dao,
		msg:         msg,
		lastFlushAt: time.Now(),
		flushEvery:  time.Second,
		flushChars:  256,
	}
}

func (f *assistantFlusher) append(delta string) {
	if delta == "" {
		return
	}
	f.temp.WriteString(delta)
}

func (f *assistantFlusher) shouldFlush() bool {
	return f.temp.Len() >= f.flushChars || time.Since(f.lastFlushAt) >= f.flushEvery
}

func (f *assistantFlusher) flush(ctx context.Context) {
	if f.temp.Len() == 0 {
		return
	}
	f.msg.Content += f.temp.String()

	_, err := f.dao.UpdateContent(ctx, f.msg)
	if err != nil {
		f.log.Errorf("update assistant content failed: session=%d message_id=%s err=%v", f.msg.SessionId, f.msg.MessageId, err)
		return
	}

	f.temp.Reset()
	f.lastFlushAt = time.Now()
}

func (f *assistantFlusher) finalize(ctx context.Context, status int64, finishReason string, usage *openaiUsage) error {
	f.msg.Status = status
	f.msg.FinishReason = finishReason
	if f.temp.Len() > 0 {
		f.msg.Content += f.temp.String()
		f.temp.Reset()
	}
	f.msg.PromptTokens = usage.PromptTokens
	f.msg.CompletionTokens = usage.CompletionTokens
	f.msg.TotalTokens = usage.TotalTokens

	var err error
	switch status {
	case MessageStatusSuccess:
		_, err = f.dao.UpdateContentFinishedSuccess(ctx, f.msg)
	case MessageStatusFailed:
		_, err = f.dao.UpdateContentFinishedFailed(ctx, f.msg)
	default:
		f.log.Errorf("错误的status")
		return errors.New("invalid status")
	}
	return err
}

type openaiUsage struct {
	PromptTokens     uint64
	CompletionTokens uint64
	TotalTokens      uint64
}
