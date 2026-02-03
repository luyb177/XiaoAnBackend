package tasks

import (
	"encoding/json"
	"fmt"
)

type VideoRelationTaskType string

const (
	// VideoRelationAdd  添加视频关联内容
	VideoRelationAdd VideoRelationTaskType = "add"

	// VideoRelationModify  更新视频关联内容
	VideoRelationModify VideoRelationTaskType = "modify"

	// VideoRelationDelete  删除视频关联内容
	VideoRelationDelete VideoRelationTaskType = "delete"

	// VideoRelationGet 获取视频
	VideoRelationGet VideoRelationTaskType = "get"
)

type VideoRelationTask struct {
	Type    VideoRelationTaskType `json:"type"`
	VideoID uint64                `json:"video_id"`
	Tags    []string              `json:"tags"`
}

func (t *VideoRelationTask) ID() string {
	return fmt.Sprintf("%s:%s:%d", VideoRelationTaskPrefix, t.Type, t.VideoID)
}

func (t *VideoRelationTask) Payload() ([]byte, error) {
	return json.Marshal(t)
}
