package convert

import (
	"github.com/luyb177/XiaoAnBackend/qa/internal/model"
	"github.com/luyb177/XiaoAnBackend/qa/internal/repo/chatmessage"
)

func CacheMsg(msg *model.ChatMessage) *chatmessage.MessageCache {
	return &chatmessage.MessageCache{
		ID:        msg.Id,
		MessageID: msg.MessageId,
		Role:      msg.Role,
		Content:   msg.Content,
		Ts:        msg.CreatedAt.Unix(),
	}
}
