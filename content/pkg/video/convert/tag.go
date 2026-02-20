package convert

import "github.com/luyb177/XiaoAnBackend/content/internal/model"

func VideoTagsFromStrings(videoID uint64, tags []string) []*model.VideoTag {
	res := make([]*model.VideoTag, len(tags))
	for i, tag := range tags {
		res[i] = &model.VideoTag{
			Tag:     tag,
			VideoID: videoID,
		}
	}
	return res
}

func StringsFromVideoTags(tags []*model.VideoTag) []string {
	res := make([]string, len(tags))
	for i, tag := range tags {
		res[i] = tag.Tag
	}
	return res
}
