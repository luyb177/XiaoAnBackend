package logic

import (
	"context"
	"fmt"

	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	TypeVideo   = "video"
	TypeComic   = "comic"
	TypePodcast = "podcast"
	TypeArticle = "article"
	TypeComment = "comment"
)

type SearchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	videoDao   model.VideoModel
	comicDao   model.ComicModel
	podcastDao model.PodcastModel
	articleDao model.ArticleModel
}

func NewSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchLogic {
	return &SearchLogic{
		ctx:        ctx,
		svcCtx:     svcCtx,
		Logger:     logx.WithContext(ctx),
		videoDao:   model.NewVideoModel(svcCtx.Mysql),
		comicDao:   model.NewComicModel(svcCtx.Mysql),
		podcastDao: model.NewPodcastModel(svcCtx.Mysql),
		articleDao: model.NewArticleModel(svcCtx.Mysql),
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

	switch in.Type {
	case TypeVideo:
		return l.SearchVideo(in)
	case TypeComic:
		return l.SearchComic(in)
	case TypePodcast:
		return l.SearchPodcast(in)
	case TypeArticle:
		return l.SearchArticle(in)
	default:
		return &v1.Response{
			Code:    400,
			Message: "请选择正确的搜索类型",
		}, fmt.Errorf("请选择正确的搜索类型")
	}
}

func (l *SearchLogic) SearchVideo(in *v1.SearchRequest) (*v1.Response, error) {
	offest := (in.Page - 1) * in.PageSize
	video, err := l.videoDao.FindByVideoTagsAndKeyWord(l.ctx, int(offest), int(in.PageSize), in.Tag, in.Keyword)
	if err != nil {
		return nil, err
	}
	videoRes := make([]*v1.Video, 0, len(video))
	// 转换
	for i := 0; i < len(video); i++ {
		videoRes = append(videoRes, &v1.Video{
			Id:           video[i].Id,
			Name:         video[i].Name,
			Url:          video[i].Url,
			Description:  video[i].Description.String,
			Cover:        video[i].Cover,
			Author:       video[i].Author,
			CreateTime:   video[i].PublishedAt.Time.Unix(),
			CreatedAt:    video[i].CreatedAt.Unix(),
			UpdateTime:   video[i].UpdatedAt.Unix(),
			LikeCount:    video[i].LikeCount,
			ViewCount:    video[i].ViewCount,
			CollectCount: video[i].CollectCount,
		})
	}

	res := &v1.SearchResponse{
		Videos: videoRes,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: "消息体类型转换失败",
		}, err
	}

	return &v1.Response{
		Code:    200,
		Message: "查询成功",
		Data:    resAny,
	}, nil
}

func (l *SearchLogic) SearchComic(in *v1.SearchRequest) (*v1.Response, error) {
	offest := (in.Page - 1) * in.PageSize
	comic, err := l.comicDao.FindByTagsAndKeyWord(l.ctx, int(offest), int(in.PageSize), in.Tag, in.Keyword)
	if err != nil {
		return nil, err
	}
	comicRes := make([]*v1.Comic, 0, len(comic))
	// 转换
	for i := 0; i < len(comic); i++ {
		comicRes = append(comicRes, &v1.Comic{
			Id:           comic[i].Id,
			Name:         comic[i].Name,
			Description:  comic[i].Description.String,
			Cover:        comic[i].Cover,
			Author:       comic[i].Author,
			CreateTime:   comic[i].PublishedAt.Unix(),
			CreatedAt:    comic[i].CreatedAt.Unix(),
			UpdateTime:   comic[i].UpdatedAt.Unix(),
			LikeCount:    comic[i].LikeCount,
			ViewCount:    comic[i].ViewCount,
			CollectCount: comic[i].CollectCount,
		})
	}

	res := &v1.SearchResponse{
		Comics: comicRes,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: "消息体类型转换失败",
		}, err
	}
	return &v1.Response{
		Code:    200,
		Message: "查询成功",
		Data:    resAny,
	}, nil
}

func (l *SearchLogic) SearchPodcast(in *v1.SearchRequest) (*v1.Response, error) {
	offest := (in.Page - 1) * in.PageSize
	podcast, err := l.podcastDao.FindByTagsAndKeyWord(l.ctx, int(offest), int(in.PageSize), in.Tag, in.Keyword)
	if err != nil {
		return nil, err
	}
	podcastRes := make([]*v1.Podcast, 0, len(podcast))
	// 转换
	for i := 0; i < len(podcast); i++ {
		podcastRes = append(podcastRes, &v1.Podcast{
			Id:          podcast[i].Id,
			Name:        podcast[i].Name,
			Description: podcast[i].Description.String,
			Cover:       podcast[i].Cover,
			Author:      podcast[i].Author,
			CreateTime:  podcast[i].PublishedAt.Time.Unix(),
			CreatedAt:   podcast[i].CreatedAt.Unix(),
			UpdateTime:  podcast[i].UpdatedAt.Unix(),
			LikeCount:   podcast[i].LikeCount,
		})
	}

	res := &v1.SearchResponse{
		Podcasts: podcastRes,
	}
	resAny, err := anypb.New(res)
	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: "消息体类型转换失败",
		}, err
	}
	return &v1.Response{
		Code:    200,
		Message: "查询成功",
		Data:    resAny,
	}, nil
}

func (l *SearchLogic) SearchArticle(in *v1.SearchRequest) (*v1.Response, error) {
	offest := (in.Page - 1) * in.PageSize
	article, err := l.articleDao.FindByTagsAndKeyWord(l.ctx, int(offest), int(in.PageSize), in.Tag, in.Keyword)
	if err != nil {
		return nil, err
	}
	articleRes := make([]*v1.Article, 0, len(article))
	// 转换
	for i := 0; i < len(article); i++ {
		articleRes = append(articleRes, &v1.Article{
			Id:           article[i].Id,
			Name:         article[i].Name,
			Description:  article[i].Description.String,
			Cover:        article[i].Cover,
			Author:       article[i].Author,
			PublishedAt:  article[i].PublishedAt.Unix(),
			CreatedAt:    article[i].CreatedAt.Unix(),
			UpdatedAt:    article[i].UpdatedAt.Unix(),
			LikeCount:    article[i].LikeCount,
			ViewCount:    article[i].ViewCount,
			CollectCount: article[i].CollectCount,
		})
	}

	res := &v1.SearchResponse{
		Articles: articleRes,
	}
	resAny, err := anypb.New(res)
	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: "消息体类型转换失败",
		}, err
	}
	return &v1.Response{
		Code:    200,
		Message: "查询成功",
		Data:    resAny,
	}, nil
}
