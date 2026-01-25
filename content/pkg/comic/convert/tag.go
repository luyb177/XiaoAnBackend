package convert

import (
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
)

func ComicTagsFromStrings(comicID uint64, tags []string) []*model.ComicTag {
	res := make([]*model.ComicTag, len(tags))
	for i, tag := range tags {
		res[i] = &model.ComicTag{
			Tag:     tag,
			ComicId: comicID,
		}
	}
	return res
}

func StringsFromComicTags(tags []*model.ComicTag) []string {
	res := make([]string, len(tags))
	for i, tag := range tags {
		res[i] = tag.Tag
	}
	return res
}
