package logic

import (
	"context"
	"fmt"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"

	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	SearchTypeVideo   = "video"
	SearchTypeComic   = "comic"
	SearchTypePodcast = "podcast"
	SearchTypeArticle = "article"
)

type SearchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	videoDao   model.VideoModel
	comicDao   model.ComicModel
	podcastDao model.PodcastModel
}

func NewSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchLogic {
	return &SearchLogic{
		ctx:        ctx,
		svcCtx:     svcCtx,
		Logger:     logx.WithContext(ctx),
		videoDao:   model.NewVideoModel(svcCtx.Mysql),
		comicDao:   model.NewComicModel(svcCtx.Mysql),
		podcastDao: model.NewPodcastModel(svcCtx.Mysql),
	}
}

// Search 搜索
func (l *SearchLogic) Search(in *v1.SearchRequest) (*v1.Response, error) {
	creatorId := l.ctx.Value("user_id").(uint64)
	creatorRole := l.ctx.Value("user_role").(string)
	creatorStatus := l.ctx.Value("user_status").(int64)

	if creatorId == 0 || creatorRole == "" || creatorStatus != 1 {
		return &v1.Response{
			Code:    400,
			Message: "用户信息错误",
		}, fmt.Errorf("用户信息错误")
	}

	if in.Keyword == "" {
		return &v1.Response{
			Code:    400,
			Message: "请输入搜索关键词",
		}, fmt.Errorf("请输入搜索关键词")
	}

	if in.Type == "" {
		return &v1.Response{
			Code:    400,
			Message: "请选择搜索类型",
		}, fmt.Errorf("请选择搜索类型")
	}

	if in.Page < 0 {
		in.Page = 1
	}
	if in.PageSize < 0 {
		in.PageSize = 10
	}

	// 查询
	var videos []*model.Video
	var comics []*model.Comic
	var podcasts []*model.Podcast

	// 偏移量
	offest := (in.Page - 1) * in.PageSize
	var err error

	switch in.Type {
	case SearchTypeVideo:
		videos, err = l.videoDao.FindByVideoTagsAndKeyWord(l.ctx, int(offest), int(in.PageSize), in.Tag, in.Keyword)
	case SearchTypeComic:
		comics, err = l.comicDao.FindByTagsAndKeyWord(l.ctx, int(offest), int(in.PageSize), in.Tag, in.Keyword)
	case SearchTypePodcast:
		podcasts, err = l.podcastDao.FindByTagsAndKeyWord(l.ctx, int(offest), int(in.PageSize), in.Tag, in.Keyword)
	}

	return &v1.Response{
		Code:    200,
		Message: "搜索成功",
	}, nil
}
