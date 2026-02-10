package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
)

type UnCollectLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnCollectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnCollectLogic {
	return &UnCollectLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UnCollect 取消收藏
func (l *UnCollectLogic) UnCollect(in *v1.UnCollectRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == InvalidUserID || user.Status != UserStatusNormal {
		return bad("用户未登录或状态异常"), nil
	}

	if resp := ValidateCollectAndUnCollectRequest(in.ContentType, in.ContentId); resp != nil {
		return resp, nil
	}

	unCollected, err := l.svcCtx.CollectRepo.UnCollect(l.ctx, user.UID, in.ContentType, in.ContentId)
	if err != nil {
		l.Errorf("UnCollectLogic UnCollect err: %v", err)
		return bad("取消收藏失败"), nil
	}

	if !unCollected {
		// 幂等处理， 用户未收藏、或者已经取消收藏
		return &v1.Response{
			Code:    200,
			Message: "取消收藏成功",
		}, nil
	}

	collectRelationTask := &tasks.CollectRelationTask{
		Type:        tasks.CollectRelationDelete,
		ContentType: in.ContentType,
		ContentID:   in.ContentId,
		UID:         user.UID,
	}

	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, collectRelationTask)
	if err != nil {
		l.Errorf("UnCollectLogic Enqueue collectRelationTask err: %v", err)
		// 任务入列失败，不影响用户操作结果
	}

	return &v1.Response{
		Code:    200,
		Message: "取消收藏成功",
	}, nil
}
