package convert

import (
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	v1 "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
)

func PBFromArticle(list []*model.Article) []*v1.Article {
	res := make([]*v1.Article, len(list))
	for i, article := range list {
		res[i] = &v1.Article{
			Id:             article.Id,
			Name:           article.Name,
			Tag:            nil,
			Url:            article.Url,
			Description:    article.Description.String,
			Cover:          article.Cover,
			Content:        article.Content.String,
			Author:         article.Author,
			PublishedAt:    article.PublishedAt.Unix(),
			CreatedAt:      article.CreatedAt.Unix(),
			UpdatedAt:      article.UpdatedAt.Unix(),
			LikeCount:      article.LikeCount,
			ViewCount:      article.ViewCount,
			CollectCount:   article.CollectCount,
			LastModifiedBy: article.LastModifiedBy.Int64,
			RelationStatus: article.RelationStatus,
			CommentCount:   article.CommentCount,
			IsLiked:        false,
			IsCollected:    false,
		}
	}
	return res
}
