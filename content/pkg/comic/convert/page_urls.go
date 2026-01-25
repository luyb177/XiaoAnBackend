package convert

import (
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	v1 "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
)

func ComicPagesFromStrings(chapterId uint64, pageUrls []string) []*model.ComicPage {
	res := make([]*model.ComicPage, len(pageUrls))
	for i, url := range pageUrls {
		res[i] = &model.ComicPage{
			ChapterId: chapterId,
			PageNo:    int64(i + 1),
			Url:       url,
		}
	}
	return res
}

func PBFromComicPages(pages []*model.ComicPage) []*v1.ComicPage {
	res := make([]*v1.ComicPage, len(pages))
	for i, page := range pages {
		res[i] = &v1.ComicPage{
			Id:             page.Id,
			ComicChapterId: page.ChapterId,
			PageNo:         page.PageNo,
			Url:            page.Url,
			CreatedAt:      page.CreatedAt.Unix(),
			UpdatedAt:      page.UpdatedAt.Unix(),
		}
	}
	return res
}
