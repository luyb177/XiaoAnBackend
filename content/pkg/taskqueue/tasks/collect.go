package tasks

import (
	"encoding/json"
	"fmt"
)

type CollectRelationTaskType string

const (
	// CollectRelationAdd 添加收藏关联内容
	CollectRelationAdd CollectRelationTaskType = "add"

	// CollectRelationDelete 删除收藏关联内容
	CollectRelationDelete CollectRelationTaskType = "delete"
)

type CollectRelationTask struct {
	Type        CollectRelationTaskType `json:"type"`         // 任务类型（添加或删除收藏关联内容）
	ContentType string                  `json:"content_type"` // 收藏类型（文章、视频等）
	ContentID   uint64                  `json:"content_id"`
	UID         uint64                  `json:"uid"`
}

func (t *CollectRelationTask) ID() string {
	return fmt.Sprintf("%s:%s:%s:%d:%d", CollectRelationTaskPrefix, t.Type, t.ContentType, t.ContentID, t.UID)
}

func (t *CollectRelationTask) Payload() ([]byte, error) {
	return json.Marshal(t)
}
