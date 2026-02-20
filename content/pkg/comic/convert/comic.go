package convert

import (
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	v1 "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
)

func PBFromComics(list []*model.Comic) []*v1.Comic {
	res := make([]*v1.Comic, len(list))
	for i, item := range list {
		res[i] = &v1.Comic{
			Id:             item.Id,
			Name:           item.Name,
			Tag:            nil,
			Description:    item.Description.String,
			Cover:          item.Cover,
			Author:         item.Author,
			PublishedAt:    item.PublishedAt.Unix(),
			CreatedAt:      item.CreatedAt.Unix(),
			UpdatedAt:      item.UpdatedAt.Unix(),
			LikeCount:      item.LikeCount,
			ViewCount:      item.ViewCount,
			CollectCount:   item.CollectCount,
			ChapterCount:   item.ChapterCount,
			CommentCount:   item.CommentCount,
			LastModifiedBy: item.LastModifiedBy.Int64,
			RelationStatus: item.RelationStatus,
			IsLiked:        false,
			IsCollected:    false,
		}
	}
	return res
}
