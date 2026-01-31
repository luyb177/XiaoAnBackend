package tasks

import (
	"encoding/json"
	"fmt"
)

type CommentRelationTaskType string

const (
	// CommentRelationAdd 添加评论关联内容
	CommentRelationAdd CommentRelationTaskType = "add"

	// CommentRelationDelete 删除评论关联内容
	CommentRelationDelete CommentRelationTaskType = "delete"
)

type CommentRelationTask struct {
	Type           CommentRelationTaskType `json:"type"`
	ContentType    string                  `json:"content_type"`     // 评论类型（文章、视频等）
	ContentID      uint64                  `json:"content_id"`       // 内容ID
	UID            uint64                  `json:"uid"`              // 用户ID
	CommentID      uint64                  `json:"comment_id"`       // 评论ID
	ParentID       uint64                  `json:"parent_id"`        // 父评论ID
	ReplyCommentID uint64                  `json:"reply_comment_id"` // 回复的评论ID
	ReplyUserID    uint64                  `json:"reply_user_id"`    // 回复的用户ID
}

// 实现 Task 接口

// ID 返回任务 ID
func (t *CommentRelationTask) ID() string {
	return fmt.Sprintf("%s:%s:%d:%d:%d", CommentRelationTaskPrefix, t.Type, t.ContentID, t.UID, t.CommentID)
}

// Payload 返回任务内容
func (t *CommentRelationTask) Payload() ([]byte, error) {
	return json.Marshal(t)
}
