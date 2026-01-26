package convert

import (
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
)

func PodcastTagsFromStrings(podcastID uint64, tags []string) []*model.PodcastTag {
	res := make([]*model.PodcastTag, len(tags))
	for i, tag := range tags {
		res[i] = &model.PodcastTag{
			Tag:       tag,
			PodcastId: podcastID,
		}
	}
	return res
}

func StringsFromPodcastTags(tags []*model.PodcastTag) []string {
	res := make([]string, len(tags))
	for i, tag := range tags {
		res[i] = tag.Tag
	}
	return res
}
