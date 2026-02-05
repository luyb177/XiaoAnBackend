package redisqueue

import (
	"encoding/json"
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

func (t *RawTask) Payload() ([]byte, error) {
	return json.Marshal(t)
}
