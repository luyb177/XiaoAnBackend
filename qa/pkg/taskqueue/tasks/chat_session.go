package tasks

import (
	"encoding/json"
	"fmt"
)

type ChatSessionTaskType string

const (
	// ChatSessionUpdateTitle 更新标题
	ChatSessionUpdateTitle ChatSessionTaskType = "update_title"
)

type ChatSessionTask struct {
	Type        ChatSessionTaskType `json:"type"`
	SessionID   uint64              `json:"session_id"`
	UserMessage string              `json:"user_message"`
}

// 实现 Task 接口

// ID 返回任务 ID
func (t *ChatSessionTask) ID() string {
	return fmt.Sprintf("%s:%s:%d", ChatSessionTaskPrefix, t.Type, t.SessionID)
}

// Payload 返回任务内容
func (t *ChatSessionTask) Payload() ([]byte, error) {
	return json.Marshal(t)
}
