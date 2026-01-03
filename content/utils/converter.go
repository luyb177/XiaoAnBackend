package utils

import (
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
)

// 转换器

func ArticleTagsFromStrings(articleID uint64, tags []string) []*model.ArticleTag {
	res := make([]*model.ArticleTag, len(tags))
	for i, tag := range tags {
		res[i] = &model.ArticleTag{
			Tag:       tag,
			ArticleId: articleID,
		}
	}
	return res
}

func StringsFromArticleTags(tags []*model.ArticleTag) []string {
	res := make([]string, len(tags))
	for i, tag := range tags {
		res[i] = tag.Tag
	}
	return res
}

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
