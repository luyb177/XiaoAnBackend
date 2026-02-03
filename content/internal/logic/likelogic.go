package logic

import (
	"context"

	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"

	"github.com/zeromicro/go-zero/core/logx"
)

type LikeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	contentLikeDao model.ContentLikeModel
}

func NewLikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeLogic {
	return &LikeLogic{
		ctx:            ctx,
		svcCtx:         svcCtx,
		Logger:         logx.WithContext(ctx),
		contentLikeDao: model.NewContentLikeModel(svcCtx.Mysql),
	}
}

// Like 点赞
func (l *LikeLogic) Like(in *v1.LikeRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == InvalidUserID || user.Status != UserStatusNormal {
		return bad("用户未登录或状态异常"), nil
	}

	if resp := ValidateContentTypeAndIDRequest(in.ContentType, in.ContentId); resp != nil {
		return resp, nil
	}

	liked, err := l.svcCtx.LikeRepo.Like(l.ctx, user.UID, in.ContentType, in.ContentId)
	if err != nil {
		l.Errorf("Like err: 点赞失败, %v", err)
		return bad("点赞失败"), nil
	}
	if !liked {
		// 幂等处理，用户之前已经点过赞了
		return &v1.Response{
			Code:    200,
			Message: "你已经点过赞了",
		}, nil
	}

	// 进入队列
	likeRelationTask := &tasks.LikeRelationTask{
		Type:        tasks.LikeRelationAdd,
		ContentType: in.ContentType,
		ContentID:   in.ContentId,
		UID:         user.UID,
	}

	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, likeRelationTask)
	if err != nil {
		l.Errorf("Like err: 点赞关系入队列失败, %v", err)
	}

	return &v1.Response{
		Code:    200,
		Message: "点赞成功",
	}, nil
}
