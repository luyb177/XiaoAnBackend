package convert

import (
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	v1 "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
)

func PBFromComicChapter(chapters []*model.ComicChapter) []*v1.ComicChapter {
	res := make([]*v1.ComicChapter, len(chapters))
	for i, chapter := range chapters {
		res[i] = &v1.ComicChapter{
			Id:          chapter.Id,
			ComicId:     chapter.ComicId,
			ChapterNo:   chapter.ChapterNo,
			Title:       chapter.Title,
			Description: chapter.Description.String,
			PageCount:   chapter.PageCount,
			Status:      chapter.Status,
			PublishedAt: chapter.PublishedAt.Unix(),
			CreatedAt:   chapter.CreatedAt.Unix(),
			UpdatedAt:   chapter.UpdatedAt.Unix(),
		}
	}
	return res
}
