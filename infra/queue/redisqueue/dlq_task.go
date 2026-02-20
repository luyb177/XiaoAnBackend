package redisqueue

import (
	"encoding/json"
	"time"
)

type DLQReason string

const (
	DLQReasonUnmarshal DLQReason = "unmarshal_failed"
	DLQReasonMarshal   DLQReason = "marshal_failed"
	DLQReasonMaxRetry  DLQReason = "max_retry_exceeded"
	DLQReasonUnknown   DLQReason = "unknown_error"
)

type DLQStage string

const (
	DLQStageDequeue DLQStage = "dequeue"
	DLQStageRetry   DLQStage = "retry"
	DLQStageUnknown DLQStage = "unknown"
)

type DLQTask struct {
	Reason     DLQReason       `json:"reason"` // 失败的原因是什么
	Stage      DLQStage        `json:"stage"`  // 在那个阶段失败
	Error      string          `json:"error"`
	FailedAt   int64           `json:"failed_at"`
	RawPayload json.RawMessage `json:"raw_payload"`
}

func NewDLQTask(reason DLQReason, stage DLQStage, err string, rawPayload json.RawMessage) *DLQTask {
	return &DLQTask{
		Reason:     reason,
		Stage:      stage,
		Error:      err,
		FailedAt:   time.Now().Unix(),
		RawPayload: rawPayload,
	}
}
