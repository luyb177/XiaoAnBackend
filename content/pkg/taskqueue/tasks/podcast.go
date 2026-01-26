package tasks

import (
	"encoding/json"
	"fmt"

	v1 "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
)

type PodcastRelationTaskType string

const (
	// PodcastRelationAdd 添加播客关联内容
	PodcastRelationAdd PodcastRelationTaskType = "add"

	// PodcastRelationModify 修改播客关联内容
	PodcastRelationModify PodcastRelationTaskType = "modify"

	// PodcastRelationDelete 删除播客关联内容
	PodcastRelationDelete PodcastRelationTaskType = "delete"
)

type PodcastRelationTask struct {
	Type       PodcastRelationTaskType `json:"type"`
	PodcastID  uint64                  `json:"podcast_id"`
	Tags       []string                `json:"tags"`
	Highlights []*v1.PodcastHighlight  `json:"highlights"`
}

func (t *PodcastRelationTask) ID() string {
	return fmt.Sprintf("%s:%s:%d", PodcastRelationTaskPrefix, t.Type, t.PodcastID)
}

// Payload 返回任务内容
func (t *PodcastRelationTask) Payload() ([]byte, error) {
	return json.Marshal(t)
}
