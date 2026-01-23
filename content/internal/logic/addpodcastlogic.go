package logic

import (
	"context"
	"database/sql"
	"time"

	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"
)

type AddPodcastLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	PodCastDao model.PodcastModel
}

func NewAddPodcastLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddPodcastLogic {
	return &AddPodcastLogic{
		ctx:        ctx,
		svcCtx:     svcCtx,
		Logger:     logx.WithContext(ctx),
		PodCastDao: model.NewPodcastModel(svcCtx.Mysql),
	}
}

// AddPodcast 添加播客
func (l *AddPodcastLogic) AddPodcast(in *v1.AddPodcastRequest) (*v1.Response, error) {
	// 添加播客只有 超级管理员 和 员工 才能添加
	user := middleware.MustGetUser(l.ctx)
	if user.UID == InvalidUserID || (user.Role != SUPERADMIN && user.Role != STAFF) || user.Status != UserStatusNormal {
		l.Errorf("AddPodcast err: 用户未登录或无权限")

		return &v1.Response{
			Code:    400,
			Message: "用户未登录或无权限",
		}, nil
	}

	// 检验请求体内容
	if in.Name == "" {
		l.Errorf("AddPodcast err: 播客名称为空")

		return &v1.Response{
			Code:    400,
			Message: "播客名称为空",
		}, nil
	}
	if in.Url == "" {
		l.Errorf("AddPodcast err: 播客链接为空")

		return &v1.Response{
			Code:    400,
			Message: "播客链接为空",
		}, nil
	}
	if in.Description == "" {
		l.Errorf("AddPodcast err: 播客描述为空")

		return &v1.Response{
			Code:    400,
			Message: "播客描述为空",
		}, nil
	}
	if in.Cover == "" {
		l.Errorf("AddPodcast err: 播客封面为空")

		return &v1.Response{
			Code:    400,
			Message: "播客封面为空",
		}, nil
	}
	if in.Author == "" {
		l.Errorf("AddPodcast err: 播客作者为空")

		return &v1.Response{
			Code:    400,
			Message: "播客作者为空",
		}, nil
	}
	now := time.Now()
	if in.PublishedAt <= 0 {
		in.PublishedAt = now.Unix()
	}
	if in.Channel == "" {
		in.Channel = "默认频道"
	}
	if in.Tags == nil {
		in.Tags = []string{"默认标签"}
	}
	if len(in.Tags) > 10 {
		l.Errorf("AddPodcast err: 标签数量超出限制")

		return &v1.Response{
			Code:    400,
			Message: "标签数量超出限制",
		}, nil
	}
	if in.Highlights == nil {
		in.Highlights = []*v1.PodcastHighlight{}
	}
	for _, v := range in.Highlights {
		if v.Second < 0 {
			return &v1.Response{
				Code:    400,
				Message: "highlight 时间为负值",
			}, nil
		}
		if v.Highlight == "" {
			return &v1.Response{
				Code:    400,
				Message: "hightlight 重点未知",
			}, nil
		}
	}
	if in.Status != PodcastStatusPublished && in.Status != PodcastStatusDraft {
		in.Status = PodcastStatusDraft
	}

	// 添加播客
	// 1. 构造
	podcast := &model.Podcast{
		Name:           in.Name,
		Url:            in.Url,
		Description:    sql.NullString{String: in.Description, Valid: true},
		Cover:          in.Cover,
		Author:         in.Author,
		PublishedAt:    sql.NullTime{Time: time.Unix(in.PublishedAt, 0), Valid: true},
		RelationStatus: RelationStatusPending,
		LastModifiedBy: sql.NullInt64{Int64: int64(user.UID), Valid: true},
		LikeCount:      0,
		ViewCount:      0,
		CollectCount:   0,
		Channel:        in.Channel,
		Status:         in.Status,
	}

	// 2. 写入数据库
	result, err := l.PodCastDao.Insert(l.ctx, podcast)
	if err != nil {
		l.Errorf("AddPodcast err: 播客添加失败，%v", err)

		return &v1.Response{
			Code:    500,
			Message: "播客添加失败",
		}, nil
	}

	// 3. 获取插入的播客ID
	podcastId, err := result.LastInsertId()
	if err != nil {
		l.Errorf("AddPodcast err: 获取播客ID失败，%v", err)

		return &v1.Response{
			Code:    500,
			Message: "获取播客ID失败",
		}, nil
	}
	podcast.Id = uint64(podcastId)

	// 4. 添加标签
	// 5. 添加重要时间点
	podcastRelationTask := &tasks.PodcastRelationTask{
		Type:       tasks.PodcastRelationAdd,
		PodcastID:  podcast.Id,
		Tags:       in.Tags,
		Highlights: in.Highlights,
	}
	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, podcastRelationTask)
	if err != nil {
		l.Errorf("AddPodcast Enqueue err: %v", err)
	}

	// 构造返回结果
	res := &v1.AddPodcastResponse{
		Id:             podcast.Id,
		RelationStatus: RelationStatusPending,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("AddPodcast err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "添加播客失败",
		}, nil
	}

	return &v1.Response{
		Code:    200,
		Message: "添加播客成功",
		Data:    resAny,
	}, nil
}
