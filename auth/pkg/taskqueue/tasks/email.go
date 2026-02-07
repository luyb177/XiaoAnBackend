package tasks

import (
	"encoding/json"
	"fmt"
)

type EmailRelationTaskType string

const (
	// EmailRelationSend 发送邮件
	EmailRelationSend EmailRelationTaskType = "send"

	// EmailCodeRelationDelete 删除验证码
	EmailCodeRelationDelete EmailRelationTaskType = "delete"
)

type EmailRelationTask struct {
	Type EmailRelationTaskType `json:"type"`
	To   string                `json:"to"`
	Code string                `json:"code"`
}

// 实现 Task 接口

// ID 返回任务 ID
func (t *EmailRelationTask) ID() string {
	return fmt.Sprintf("%s:%s:%s:%s", EmailRelationTaskPrefix, t.Type, t.To, t.Code)
}

// Payload 返回任务内容
func (t *EmailRelationTask) Payload() ([]byte, error) {
	return json.Marshal(t)
}
