package logic

import (
	"context"
	"errors"

	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/podcast/convert"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"
)

type GetPodcastLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	PodcastDao          model.PodcastModel
	PodcastTagDao       model.PodcastTagModel
	PodcastHighlightDao model.PodcastHighlightModel
}

func NewGetPodcastLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPodcastLogic {
	return &GetPodcastLogic{
		ctx:                 ctx,
		svcCtx:              svcCtx,
		Logger:              logx.WithContext(ctx),
		PodcastDao:          model.NewPodcastModel(svcCtx.Mysql),
		PodcastTagDao:       model.NewPodcastTagModel(svcCtx.Mysql),
		PodcastHighlightDao: model.NewPodcastHighlightModel(svcCtx.Mysql),
	}
}

// GetPodcast 获取播客
func (l *GetPodcastLogic) GetPodcast(in *v1.GetPodcastRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == InvalidUserID || user.Status != UserStatusNormal {
		return bad("用户未登录或登录状态异常"), nil
	}

	if in.Id == 0 {
		return bad("播客ID不合法"), nil
	}

	// 1. 获取播客主体信息
	podcast, err := l.PodcastDao.FindOneWithNotDelete(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Errorf("GetPodcast err: 播客不存在")

			return &v1.Response{
				Code:    404,
				Message: "播客不存在",
			}, nil
		}
		l.Errorf("GetPodcast err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "系统内部错误",
		}, nil
	}

	// 2. 异步获取 tag like
	type TagResult struct {
		podcastTags []*model.PodcastTag
		err         error
	}

	type LikeResult struct {
		liked bool
		err   error
	}
	type HighlightResult struct {
		highlights []*model.PodcastHighlight
		err        error
	}

	tagCh := make(chan TagResult, 1)
	likeCh := make(chan LikeResult, 1)
	highlightCh := make(chan HighlightResult, 1)

	go func() {
		tags, err := l.PodcastTagDao.FindManyByPodcastId(l.ctx, in.Id)
		tagCh <- TagResult{
			podcastTags: tags,
			err:         err,
		}
	}()

	go func() {
		liked, err := l.svcCtx.LikeRepo.HasLiked(l.ctx, user.UID, ContentTypePodcast, in.Id)
		likeCh <- LikeResult{
			liked: liked,
			err:   err,
		}
	}()

	go func() {
		highlights, err := l.PodcastHighlightDao.FindManyByPodcastId(l.ctx, in.Id)
		highlightCh <- HighlightResult{
			highlights: highlights,
			err:        err,
		}
	}()

	tagResult := <-tagCh
	if tagResult.err != nil {
		l.Errorf("GetPodcast err: 获取tag错误 %v", tagResult.err)
		// 不影响获取播客内容
	}
	likeResult := <-likeCh
	if likeResult.err != nil {
		l.Errorf("GetPodcast err: 获取点赞情况错误 %v", likeResult.err)
	}
	highlightResult := <-highlightCh
	if highlightResult.err != nil {
		l.Errorf("GetPodcast err: 获取重点时间点错误 %v", highlightResult.err)
		// 不影响获取播客内容
	}

	tagsRes := convert.StringsFromPodcastTags(tagResult.podcastTags)
	hightlightsRes := convert.PBFromPodcastHighlights(highlightResult.highlights)

	// 4. 构造相应
	res := &v1.GetPodcastResponse{Podcast: &v1.Podcast{
		Id:             podcast.Id,
		Name:           podcast.Name,
		Tag:            tagsRes,
		Url:            podcast.Url,
		Description:    podcast.Description.String,
		Cover:          podcast.Cover,
		Author:         podcast.Author,
		PublishedAt:    podcast.PublishedAt.Time.Unix(),
		CreatedAt:      podcast.CreatedAt.Unix(),
		UpdatedAt:      podcast.UpdatedAt.Unix(),
		LikeCount:      podcast.LikeCount,
		ViewCount:      podcast.ViewCount,
		CollectCount:   podcast.CollectCount,
		Highlights:     hightlightsRes,
		LastModifiedBy: podcast.LastModifiedBy.Int64,
		RelationStatus: podcast.RelationStatus,
		Channel:        podcast.Channel,
		Status:         podcast.Status,
		CommentCount:   podcast.CommentCount,
		IsLiked:        likeResult.liked,
	}}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("GetPodcast err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "系统内部错误",
		}, nil
	}

	msg := "获取播客成功"
	if podcast.RelationStatus == RelationStatusPending {
		msg = "播客相关内容同步中"
	}
	return &v1.Response{
		Code:    200,
		Message: msg,
		Data:    resAny,
	}, nil
}
