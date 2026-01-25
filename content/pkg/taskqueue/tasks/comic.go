package tasks

import (
	"encoding/json"
	"fmt"
)

type ComicRelationTaskType string

const (
	// ComicRelationAdd 添加漫画关联内容
	ComicRelationAdd ComicRelationTaskType = "add"

	// ComicRelationModify 修改漫画关联内容
	ComicRelationModify ComicRelationTaskType = "modify"

	// ComicRelationDelete 删除漫画关联内容
	ComicRelationDelete ComicRelationTaskType = "delete"
)

type ComicRelationTask struct {
	Type    ComicRelationTaskType `json:"type"`
	ComicID uint64                `json:"comic_id"`
	UID     uint64                `json:"uid"`
	Tags    []string              `json:"tags"`
}

// 实现 Task 接口

// ID 返回任务 ID
func (t *ComicRelationTask) ID() string {
	return fmt.Sprintf("%s:%s:%d", ComicRelationTaskPrefix, t.Type, t.ComicID)
}

// Payload 返回任务内容
func (t *ComicRelationTask) Payload() ([]byte, error) {
	return json.Marshal(t)
}
