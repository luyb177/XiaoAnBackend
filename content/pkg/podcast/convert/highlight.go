package convert

import (
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	v1 "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
)

func PodcastHighlightsFromPB(podcastID uint64, highlights []*v1.PodcastHighlight) []*model.PodcastHighlight {
	res := make([]*model.PodcastHighlight, len(highlights))
	for i, h := range highlights {
		res[i] = &model.PodcastHighlight{
			PodcastId: podcastID,
			Second:    h.Second,
			Highlight: h.Highlight,
		}
	}
	return res
}

func PBFromPodcastHighlights(highlights []*model.PodcastHighlight) []*v1.PodcastHighlight {
	res := make([]*v1.PodcastHighlight, len(highlights))
	for i, h := range highlights {
		res[i] = &v1.PodcastHighlight{
			Second:    h.Second,
			Highlight: h.Highlight,
		}
	}
	return res
}
