package logic

import (
	"context"
	"errors"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"

	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/video/convert"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"
)

type GetVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	VideoDao    model.VideoModel
	VideoTagDao model.VideoTagModel
}

func NewGetVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVideoLogic {
	return &GetVideoLogic{
		ctx:         ctx,
		svcCtx:      svcCtx,
		Logger:      logx.WithContext(ctx),
		VideoDao:    model.NewVideoModel(svcCtx.Mysql),
		VideoTagDao: model.NewVideoTagModel(svcCtx.Mysql),
	}
}

// GetVideo 获取视频
func (l *GetVideoLogic) GetVideo(in *v1.GetVideoRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == InvalidUserID || user.Status != UserStatusNormal {
		return bad("用户未登录或登录状态异常"), nil
	}

	if in.Id == 0 {
		return bad("视频ID不合法"), nil
	}

	video, err := l.VideoDao.FindOneWithNotDelete(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Errorf("GetVideo err: 视频不存在")

			return &v1.Response{
				Code:    400,
				Message: "视频不存在",
			}, nil
		}

		l.Errorf("GetVideo err: %v", err)
		return internal("获取视频失败"), nil
	}

	// 入队
	videoRelationTask := &tasks.VideoRelationTask{
		Type:    tasks.VideoRelationGet,
		VideoID: in.Id,
		Tags:    nil,
	}
	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, videoRelationTask)
	if err != nil {
		l.Errorf("GetVideo err: 入队失败 %v", err)
	}

	// 异步获取tag like
	type tagResult struct {
		tags []*model.VideoTag
		err  error
	}
	type LikeResult struct {
		liked bool
		err   error
	}

	tagCh := make(chan tagResult, 1)
	likeCh := make(chan LikeResult, 1)

	go func() {
		tags, err := l.VideoTagDao.FindManyByVideoId(l.ctx, video.Id)
		tagCh <- tagResult{
			tags: tags,
			err:  err,
		}
	}()
	go func() {
		liked, err := l.svcCtx.LikeRepo.HasLiked(l.ctx, user.UID, ContentTypeVideo, in.Id)
		likeCh <- LikeResult{
			liked: liked,
			err:   err,
		}
	}()

	tagsResult := <-tagCh
	if tagsResult.err != nil {
		l.Errorf("GetVideo err: %v", tagsResult.err)
		// 不影响获取视频内容
	}
	likeResult := <-likeCh
	if likeResult.err != nil {
		l.Errorf("GetVideo err: 获取点赞情况错误 %v", likeResult.err)
	}

	// 处理 tag
	tagsRes := convert.StringsFromVideoTags(tagsResult.tags)

	// 构造返回内容
	res := &v1.GetVideoResponse{Video: &v1.Video{
		Id:           video.Id,
		Name:         video.Name,
		Tag:          tagsRes,
		Url:          video.Url,
		Description:  video.Description.String,
		Cover:        video.Cover,
		Author:       video.Author,
		PublishedAt:  video.PublishedAt.Time.Unix(),
		CreatedAt:    video.CreatedAt.Unix(),
		UpdatedAt:    video.UpdatedAt.Unix(),
		LikeCount:    video.LikeCount,
		ViewCount:    video.ViewCount,
		CollectCount: video.CollectCount,
		CommentCount: video.CommentCount,
		IsLiked:      likeResult.liked,
	}}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("GetVideo err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "获取视频失败",
		}, nil
	}

	msg := "获取视频成功"
	if video.RelationStatus == RelationStatusPending {
		msg = "视频相关内容同步中"
	}

	return &v1.Response{
		Code:    200,
		Message: msg,
		Data:    resAny,
	}, nil
}
