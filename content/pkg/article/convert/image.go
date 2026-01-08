package convert

import (
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
)

func ArticleImagesFromPB(articleID uint64, images []*content.ArticleImage) []*model.ArticleImage {
	res := make([]*model.ArticleImage, len(images))
	for i, img := range images {
		res[i] = &model.ArticleImage{
			ArticleId: articleID,
			Url:       img.Url,
			Sort:      img.Sort,
			Type:      img.Tp,
		}
	}
	return res
}

func ArticleImagesToPB(images []*model.ArticleImage) []*content.ArticleImage {
	res := make([]*content.ArticleImage, len(images))
	for i, img := range images {
		res[i] = &content.ArticleImage{
			Url:  img.Url,
			Sort: img.Sort,
			Tp:   img.Type,
		}
	}
	return res
}
