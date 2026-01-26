package tasks

import (
	"encoding/json"
	"fmt"
)

type ArticleRelationTaskType string

const (
	// ArticleRelationAdd 添加文章关联内容
	ArticleRelationAdd ArticleRelationTaskType = "add"

	// ArticleRelationModify 修改文章关联内容
	ArticleRelationModify ArticleRelationTaskType = "modify"

	// ArticleRelationDelete 删除文章关联内容
	ArticleRelationDelete ArticleRelationTaskType = "delete"
)

type ArticleRelationTask struct {
	Type      ArticleRelationTaskType `json:"type"`
	ArticleID uint64                  `json:"article_id"`
	UID       uint64                  `json:"uid"`
	Tags      []string                `json:"tags"`
}

// 实现 Task 接口

// ID 返回任务 ID
func (t *ArticleRelationTask) ID() string {
	return fmt.Sprintf("%s:%s:%d:%d", ArticleRelationTaskPrefix, t.Type, t.ArticleID, t.UID)
}

// Payload 返回任务内容
func (t *ArticleRelationTask) Payload() ([]byte, error) {
	return json.Marshal(t)
}
