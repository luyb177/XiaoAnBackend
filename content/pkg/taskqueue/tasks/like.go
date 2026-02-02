package tasks

import (
	"encoding/json"
	"fmt"
)

type LikeRelationTaskType string

const (
	// LikeRelationAdd 添加点赞关联内容
	LikeRelationAdd LikeRelationTaskType = "add"

	// LikeRelationDelete 删除点赞关联内容
	LikeRelationDelete LikeRelationTaskType = "delete"
)

type LikeRelationTask struct {
	Type        LikeRelationTaskType `json:"type"`
	ContentType string               `json:"content_type"` // 点赞类型（文章、视频等）
	ContentID   uint64               `json:"content_id"`
	UID         uint64               `json:"uid"`
}

func (t *LikeRelationTask) ID() string {
	return fmt.Sprintf("%s:%d:%d", t.ContentType, t.ContentID, t.UID)
}

func (t *LikeRelationTask) Payload() ([]byte, error) {
	return json.Marshal(t)
}
