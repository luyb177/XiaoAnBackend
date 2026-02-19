package logic

import (
	"context"
	"database/sql"
	"time"

	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
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
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || (user.Role != constants.SUPERADMIN && user.Role != constants.STAFF) || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或状态异常"), nil
	}

	// 检验请求体内容
	validations := []Validation{
		{in.Name != "", "播客名称为空"},
		{in.Url != "", "播客链接为空"},
		{in.Description != "", "播客描述为空"},
		{in.Cover != "", "播客封面为空"},
		{in.Author != "", "播客作者为空"},
		{len(in.Tags) <= 10, "标签数量超出限制"},
	}
	for _, v := range validations {
		if !v.Condition {
			l.Errorf("AddPodcast err: %s", v.Message)

			return &v1.Response{
				Code:    400,
				Message: v.Message,
			}, nil
		}
	}

	// 设置默认值
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
	if in.Highlights == nil {
		in.Highlights = []*v1.PodcastHighlight{}
	}
	for _, v := range in.Highlights {
		if v.Highlight == "" {
			return &v1.Response{
				Code:    400,
				Message: "highlight 重点未知",
			}, nil
		}
	}
	if in.Status != PodcastStatusPublished && in.Status != PodcastStatusDraft {
		in.Status = PodcastStatusDraft
	}

	// 添加播客
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
		DeletedAt:      0,
		Channel:        in.Channel,
		Status:         in.Status,
	}

	// 写入数据库
	result, err := l.PodCastDao.Insert(l.ctx, podcast)
	if err != nil {
		l.Errorf("AddPodcast err: 播客添加失败，%v", err)

		return &v1.Response{
			Code:    500,
			Message: "播客添加失败",
		}, nil
	}

	// 获取插入的播客ID
	podcastID, err := result.LastInsertId()
	if err != nil {
		l.Errorf("AddPodcast err: 获取播客ID失败，%v", err)

		return &v1.Response{
			Code:    500,
			Message: "获取播客ID失败",
		}, nil
	}
	podcast.Id = uint64(podcastID)

	//  添加标签 & 添加重要时间点
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
