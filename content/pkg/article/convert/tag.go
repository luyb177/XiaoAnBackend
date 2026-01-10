package convert

import "github.com/luyb177/XiaoAnBackend/content/internal/model"

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
