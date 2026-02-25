package convert

import (
	"github.com/luyb177/XiaoAnBackend/qa/internal/model"
	v1 "github.com/luyb177/XiaoAnBackend/qa/pb/qa/v1"
)

func PBFromChatMessages(list []*model.ChatMessage) []*v1.ChatMessage {
	res := make([]*v1.ChatMessage, len(list))
	for i, item := range list {
		res[i] = &v1.ChatMessage{
			Id:               item.Id,
			MessageId:        item.MessageId,
			SessionId:        item.SessionId,
			UserId:           item.UserId,
			Status:           item.Status,
			PromptTokens:     item.PromptTokens,
			CompletionTokens: item.CompletionTokens,
			TotalTokens:      item.TotalTokens,
			MessageType:      item.MessageType,
			Content:          item.Content,
			FinishReason:     item.FinishReason,
			CreatedAt:        item.CreatedAt.Unix(),
			UpdatedAt:        item.UpdatedAt.Unix(),
		}
	}
	return res
}
