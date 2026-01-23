package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"

	"github.com/zeromicro/go-zero/core/logx"
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
	user := middleware.MustGetUser(l.ctx)
	if user.UID == InvalidUserID || (user.Role != SUPERADMIN && user.Role != STAFF) || user.Status != UserStatusNormal {
		l.Errorf("DeleteVideo err: 用户未登录或登录状态异常")

		return &v1.Response{
			Code:    400,
			Message: "用户未登录或登录状态异常",
		}, nil
	}

	if in.Id <= 0 {
		l.Errorf("DeleteVideo err: 视频ID错误")

		return &v1.Response{
			Code:    400,
			Message: "视频ID错误",
		}, nil
	}

	// 1. 查询视频是否存在
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

	// 2. 有 软删除
	video.DeletedAt = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	err = l.VideoDao.Update(l.ctx, video)
	if err != nil {
		l.Errorf("DeleteVideo  err: 删除视频时出错")
		return &v1.Response{
			Code:    500,
			Message: "删除视频时出错",
		}, nil
	}

	// 3. 删除相关标签
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
