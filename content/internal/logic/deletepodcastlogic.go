package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"
)

type DeletePodcastLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	PodcastDao model.PodcastModel
}

func NewDeletePodcastLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePodcastLogic {
	return &DeletePodcastLogic{
		ctx:        ctx,
		svcCtx:     svcCtx,
		Logger:     logx.WithContext(ctx),
		PodcastDao: model.NewPodcastModel(svcCtx.Mysql),
	}
}

// DeletePodcast 删除播客
func (l *DeletePodcastLogic) DeletePodcast(in *v1.DeletePodcastRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || (user.Role != constants.SUPERADMIN && user.Role != constants.STAFF) || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或状态异常"), nil
	}

	validations := []Validation{
		{in.Id > 0, "播客ID错误"},
	}
	for _, v := range validations {
		if !v.Condition {
			l.Errorf("DeletePodcast err: %s", v.Message)

			return &v1.Response{
				Code:    400,
				Message: v.Message,
			}, nil
		}
	}

	podcast, err := l.PodcastDao.FindOneWithNotDelete(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Errorf("DeletePodcast err: 播客不存在")

			return &v1.Response{
				Code:    400,
				Message: "播客不存在",
			}, nil
		}
		l.Errorf("DeletePodcast err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "服务器错误",
		}, nil
	}

	// todo 使用 soft delete
	deletedAt := uint64(time.Now().Unix())
	modifier := sql.NullInt64{Int64: int64(user.UID), Valid: true}
	err = l.PodcastDao.SoftDelete(l.ctx, podcast.Id, deletedAt, modifier)
	if err != nil {
		l.Errorf("DeletePodcast SoftDelete err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "服务器错误",
		}, nil
	}

	// 删除标签和 highlight
	podcastRelationTask := &tasks.PodcastRelationTask{
		Type:       tasks.PodcastRelationDelete,
		PodcastID:  podcast.Id,
		Tags:       nil,
		Highlights: nil,
	}
	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, podcastRelationTask)
	if err != nil {
		l.Errorf("DeletePodcast Enqueue err: %v", err)
	}

	return &v1.Response{
		Code:    200,
		Message: "删除播客成功",
	}, nil
}
