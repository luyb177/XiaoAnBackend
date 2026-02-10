package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
)

type CollectLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCollectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CollectLogic {
	return &CollectLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Collect 收藏
func (l *CollectLogic) Collect(in *v1.CollectRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == InvalidUserID || user.Status != UserStatusNormal {
		return bad("用户未登录或状态异常"), nil
	}
	if resp := ValidateCollectAndUnCollectRequest(in.ContentType, in.ContentId); resp != nil {
		return resp, nil
	}

	collected, err := l.svcCtx.CollectRepo.Collect(l.ctx, user.UID, in.ContentType, in.ContentId)
	if err != nil {
		l.Errorf("Collect err: 收藏失败, %v", err)
		return bad("收藏失败"), nil
	}
	if !collected {
		// 幂等处理，用户之前已经收藏过了
		return &v1.Response{
			Code:    200,
			Message: "你已经收藏过了",
		}, nil
	}

	// 进入队列
	collectRelationTask := &tasks.CollectRelationTask{
		Type:        tasks.CollectRelationAdd,
		ContentType: in.ContentType,
		ContentID:   in.ContentId,
		UID:         user.UID,
	}

	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, collectRelationTask)
	if err != nil {
		l.Errorf("Collect err: 收藏任务入队失败, %v", err)
	}

	return &v1.Response{
		Code:    200,
		Message: "收藏成功",
	}, nil
}
