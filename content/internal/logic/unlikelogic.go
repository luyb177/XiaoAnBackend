package logic

import (
	"context"
	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"

	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnlikeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnlikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnlikeLogic {
	return &UnlikeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Unlike 取消点赞
func (l *UnlikeLogic) Unlike(in *v1.UnlikeRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID <= 0 || user.Status != UserStatusNormal {
		return bad("用户未登录或状态异常"), nil
	}

	if resp := l.validate(in); resp != nil {
		return resp, nil
	}

	unliked, err := l.svcCtx.LikeRepo.Unlike(l.ctx, user.UID, in.ContentType, in.ContentId)
	if err != nil {
		l.Errorf("UnlikeLogic unlike err: %v", err)
		return bad("取消点赞失败"), nil
	}

	if !unliked {
		// 幂等处理， 用户未点过赞、或者已经取消点赞
		return &v1.Response{
			Code:    200,
			Message: "取消点赞成功",
		}, nil
	}

	likeRelationTask := &tasks.LikeRelationTask{
		Type:        tasks.LikeRelationDelete,
		ContentType: in.ContentType,
		ContentID:   in.ContentId,
		UID:         user.UID,
	}

	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, likeRelationTask)
	if err != nil {
		l.Errorf("Unlike err: 取消点赞关系任务入队列失败, %v", err)
	}

	return &v1.Response{
		Code:    200,
		Message: "取消点赞成功",
	}, nil
}

func (l *UnlikeLogic) validate(in *v1.UnlikeRequest) *v1.Response {
	switch {
	case in.ContentType == "":
		return bad("内容类型不能为空")
	case in.ContentType != ContentTypeArticle && in.ContentType != ContentTypePodcast &&
		in.ContentType != ContentTypeVideo && in.ContentType != ContentTypeComic &&
		in.ContentType != ContentTypeComment:
		return bad("内容类型不合法")
	case in.ContentId == 0:
		return bad("内容ID不能为0")
	}
	return nil
}
