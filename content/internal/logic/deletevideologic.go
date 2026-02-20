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

type DeleteVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	VideoDao model.VideoModel
}

func NewDeleteVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteVideoLogic {
	return &DeleteVideoLogic{
		ctx:      ctx,
		svcCtx:   svcCtx,
		Logger:   logx.WithContext(ctx),
		VideoDao: model.NewVideoModel(svcCtx.Mysql),
	}
}

// DeleteVideo 删除视频
func (l *DeleteVideoLogic) DeleteVideo(in *v1.DeleteVideoRequest) (*v1.Response, error) {
	// 目前是只有超级管理员和员工可以删除视频
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || (user.Role != constants.SUPERADMIN && user.Role != constants.STAFF) || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或状态异常"), nil
	}

	validations := []Validation{
		{in.Id > 0, "视频ID错误"},
	}
	for _, v := range validations {
		if !v.Condition {
			l.Errorf("DeleteVideo err: %s", v.Message)

			return &v1.Response{
				Code:    400,
				Message: v.Message,
			}, nil
		}
	}

	// 查询视频是否存在
	video, err := l.VideoDao.FindOneWithNotDelete(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Errorf("DeleteVideo  err: 视频不存在")

			return &v1.Response{
				Code:    400,
				Message: "视频不存在",
			}, nil
		}
		l.Errorf("DeleteVideo  err: 查询视频时出错")
		return &v1.Response{
			Code:    500,
			Message: "查询视频时出错",
		}, nil
	}

	// 有 软删除
	deletedAt := uint64(time.Now().Unix())
	modifier := sql.NullInt64{Int64: int64(user.UID), Valid: true}
	err = l.VideoDao.SoftDelete(l.ctx, video.Id, deletedAt, modifier)
	if err != nil {
		l.Errorf("DeleteVideo  err: 删除视频时出错")
		return &v1.Response{
			Code:    500,
			Message: "删除视频时出错",
		}, nil
	}

	// 删除相关标签
	videoRelationTask := &tasks.VideoRelationTask{
		Type:    tasks.VideoRelationDelete,
		VideoID: video.Id,
		Tags:    nil,
	}
	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, videoRelationTask)
	if err != nil {
		l.Errorf("DeleteVideo  Enqueue err: %v", err)
	}

	return &v1.Response{
		Code:    200,
		Message: "删除视频成功",
	}, nil
}
