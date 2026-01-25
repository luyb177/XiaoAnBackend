package tasks

import (
	"encoding/json"
	"fmt"
)

type ComicChapterRelationTaskType string

const (
	// ComicChapterRelationAdd 添加漫画章节关联内容
	ComicChapterRelationAdd ComicChapterRelationTaskType = "add"

	// ComicChapterRelationModify 修改漫画章节关联内容
	ComicChapterRelationModify ComicChapterRelationTaskType = "modify"

	// ComicChapterRelationDelete 删除漫画章节关联内容
	ComicChapterRelationDelete ComicChapterRelationTaskType = "delete"

	// ComicChapterRelationDeleteAll 删除漫画章节所有关联内容
	ComicChapterRelationDeleteAll ComicChapterRelationTaskType = "delete_all"
)

type ComicChapterRelationTask struct {
	Type      ComicChapterRelationTaskType `json:"type"`
	ComicId   uint64                       `json:"comic_id"`
	UID       uint64                       `json:"uid"`
	ChapterID uint64                       `json:"chapter_id"`
	PageUrls  []string                     `json:"page_urls"`
}

// 实现 Task 接口

// ID 返回任务 ID
func (t *ComicChapterRelationTask) ID() string {
	return fmt.Sprintf("%s:%s:%d:%d:%d", ComicChapterRelationTaskPrefix, t.Type, t.ComicId, t.ChapterID, t.UID)
}

// Payload 返回任务内容
func (t *ComicChapterRelationTask) Payload() ([]byte, error) {
	return json.Marshal(t)
}
