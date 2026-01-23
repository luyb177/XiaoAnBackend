package logic

import (
	"context"
	"errors"

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
	if in.Id <= 0 {
		l.Errorf("GetVideo err: 获取视频参数错误")

		return &v1.Response{
			Code:    400,
			Message: "参数错误",
		}, nil
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
		return &v1.Response{
			Code:    500,
			Message: "获取视频失败",
		}, nil
	}

	// 异步获取tag
	type tagResult struct {
		tags []*model.VideoTag
		err  error
	}

	tagCh := make(chan tagResult, 1)

	go func() {
		tags, err := l.VideoTagDao.FindManyByVideoId(l.ctx, video.Id)
		tagCh <- tagResult{
			tags: tags,
			err:  err,
		}
	}()

	tagsResult := <-tagCh
	if tagsResult.err != nil {
		l.Errorf("GetVideo err: %v", tagsResult.err)
		// 不影响获取视频内容
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
