package redisqueue

import (
	"encoding/json"
	"fmt"

	v1 "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue"
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

func NewArticleRelationTask(tp ArticleRelationTaskType, articleID uint64, tags []string, images []*v1.ArticleImage) taskqueue.Task {
	return &ArticleRelationTask{
		Type:      tp,
		ArticleID: articleID,
		Tags:      tags,
	}
}

// 实现 Task 接口

// ID 返回任务 ID
func (t *ArticleRelationTask) ID() string {
	return fmt.Sprintf("article_relation_task:%s:%d", t.Type, t.ArticleID)
}

// Payload 返回任务内容
func (t *ArticleRelationTask) Payload() []byte {
	b, _ := json.Marshal(t)
	return b
}
