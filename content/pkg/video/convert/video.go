package convert

import (
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	v1 "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
)

func PBFromVideo(list []*model.Video) []*v1.Video {
	res := make([]*v1.Video, len(list))
	for i, item := range list {
		res[i] = &v1.Video{
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
			CommentCount:   item.CommentCount,
			LastModifiedBy: item.LastModifiedBy.Int64,
			RelationStatus: item.RelationStatus,
			IsLiked:        false,
			IsCollected:    false,
		}
	}
	return res
}
