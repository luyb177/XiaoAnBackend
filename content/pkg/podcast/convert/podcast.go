package convert

import (
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	v1 "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
)

func PBFromPodcasts(list []*model.Podcast) []*v1.Podcast {
	res := make([]*v1.Podcast, len(list))
	for i, item := range list {
		res[i] = &v1.Podcast{
			Id:             item.Id,
			Name:           item.Name,
			Tag:            nil,
			Url:            item.Url,
			Description:    item.Description.String,
			Cover:          item.Cover,
			Author:         item.Author,
			PublishedAt:    item.PublishedAt.Time.Unix(),
			CreatedAt:      item.CreatedAt.Unix(),
			UpdatedAt:      item.UpdatedAt.Unix(),
			LikeCount:      item.LikeCount,
			ViewCount:      item.ViewCount,
			CollectCount:   item.CollectCount,
			Highlights:     nil,
			LastModifiedBy: item.LastModifiedBy.Int64,
			RelationStatus: item.RelationStatus,
			Channel:        item.Channel,
			Status:         item.Status,
			CommentCount:   item.CommentCount,
			IsLiked:        false,
			IsCollected:    false,
		}
	}
	return res
}
