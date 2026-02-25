package convert

import (
	"github.com/luyb177/XiaoAnBackend/qa/internal/model"
	v1 "github.com/luyb177/XiaoAnBackend/qa/pb/qa/v1"
)

func PBFromChatSessions(list []*model.ChatSession) []*v1.ChatSession {
	res := make([]*v1.ChatSession, len(list))
	for i, item := range list {
		res[i] = &v1.ChatSession{
			Id:             item.Id,
			Title:          item.Title,
			UserId:         item.UserId,
			MessageCount:   item.MessageCount,
			HasMessage:     item.HasMessage,
			SessionStatus:  item.SessionStatus,
			IsPinned:       item.IsPinned,
			PinnedAt:       item.PinnedAt.Time.Unix(),
			LastMessageAt:  item.LastMessageAt.Unix(),
			RelationStatus: item.RelationStatus,
			CreatedAt:      item.CreatedAt.Unix(),
			UpdatedAt:      item.UpdatedAt.Unix(),
		}
	}
	return res
}
