package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"
)

type ModifyPodcastLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	PodcastDao model.PodcastModel
}

func NewModifyPodcastLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyPodcastLogic {
	return &ModifyPodcastLogic{
		ctx:        ctx,
		svcCtx:     svcCtx,
		Logger:     logx.WithContext(ctx),
		PodcastDao: model.NewPodcastModel(svcCtx.Mysql),
	}
}

// ModifyPodcast 修改播客
func (l *ModifyPodcastLogic) ModifyPodcast(in *v1.ModifyPodcastRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || (user.Role != constants.SUPERADMIN && user.Role != constants.STAFF) || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或状态异常"), nil
	}

	// 验证参数
	validations := []Validation{
		{in.Id > 0, "播客ID不能小于等于0"},
		{in.Name != "", "播客名称为空"},
		{in.Url != "", "播客链接为空"},
		{in.Description != "", "播客描述为空"},
		{in.Cover != "", "播客封面为空"},
		{in.Author != "", "播客作者为空"},
		{len(in.Tags) <= 10, "标签数量超出限制"},
	}

	for _, v := range validations {
		if !v.Condition {
			l.Errorf("ModifyPodcast err: %s", v.Message)
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
	if len(in.Tags) == 0 {
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

	// 1. 获取播客以验证存在性
	podcast, err := l.PodcastDao.FindOneWithNotDelete(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Errorf("ModifyPodcast err: 播客不存在")

			return &v1.Response{
				Code:    404,
				Message: "播客不存在",
			}, nil
		}
		l.Errorf("ModifyPodcast err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "获取播客信息失败",
		}, nil
	}

	// 2. 更新播客主体
	podcast.Name = in.Name
	podcast.Url = in.Url
	podcast.Description = sql.NullString{String: in.Description, Valid: true}
	podcast.Cover = in.Cover
	podcast.Author = in.Author
	podcast.PublishedAt = sql.NullTime{Time: time.Unix(in.PublishedAt, 0), Valid: true}
	podcast.Channel = in.Channel
	podcast.Status = in.Status
	podcast.RelationStatus = RelationStatusPending
	podcast.LastModifiedBy = sql.NullInt64{Int64: int64(user.UID), Valid: true}
	podcast.DeletedAt = 0

	err = l.PodcastDao.Update(l.ctx, podcast)
	if err != nil {
		l.Errorf("ModifyPodcast err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "修改播客失败",
		}, nil
	}

	// 3. 修改标签和高亮
	podcastRelationTask := &tasks.PodcastRelationTask{
		Type:       tasks.PodcastRelationModify,
		PodcastID:  podcast.Id,
		Tags:       in.Tags,
		Highlights: in.Highlights,
	}
	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, podcastRelationTask)
	if err != nil {
		l.Errorf("ModifyPodcast err: 播客关联信息修改任务入列失败: %v", err)
	}

	// 4. 构造返回值
	res := &v1.ModifyPodcastResponse{
		Id:             podcast.Id,
		RelationStatus: RelationStatusPending,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("ModifyPodcast err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "系统内部错误",
		}, nil
	}

	return &v1.Response{
		Code:    200,
		Message: "修改播客成功",
		Data:    resAny,
	}, nil
}
