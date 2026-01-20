package redisqueue

import (
	"encoding/json"
	"fmt"
	"time"
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
	Tags      []string                `json:"tags"`
}

func NewArticleRelationTask(task *ArticleRelationTask) (*RawTask, error) {
	data, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()

	return &RawTask{
		TaskID:    fmt.Sprintf("article_relation_%s_%d", task.Type, task.ArticleID),
		Retry:     0,
		MaxRetry:  MaxRetry,
		DelaySec:  0,
		Data:      data,
		CreatedAt: now,
	}, nil
}

// 实现 Task 接口

// ID 返回任务 ID
func (t *ArticleRelationTask) ID() string {
	return fmt.Sprintf("article_relation_task:%s:%d", t.Type, t.ArticleID)
}

// Payload 返回任务内容
func (t *ArticleRelationTask) Payload() []byte {
	b, err := json.Marshal(t)
	if err != nil {
		panic(fmt.Sprintf("ArticleRelationTask marshal error: %v", err))
	}
	return b
}
