package logic

import (
	"context"
	"errors"
	"time"

	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
)

type DeleteCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	CommentDao model.CommentModel
}

func NewDeleteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCommentLogic {
	return &DeleteCommentLogic{
		ctx:        ctx,
		svcCtx:     svcCtx,
		Logger:     logx.WithContext(ctx),
		CommentDao: model.NewCommentModel(svcCtx.Mysql),
	}
}

// DeleteComment 删除评论
func (l *DeleteCommentLogic) DeleteComment(in *v1.DeleteCommentRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或状态异常"), nil
	}

	// 参数校验
	if in.Id <= 0 {
		l.Errorf("DeleteComment err: 参数校验失败, 评论ID无效")
		return bad("评论ID无效"), nil
	}

	// 获取一下评论
	comment, err := l.CommentDao.FindOneWithNotDeleted(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Errorf("DeleteComment err: 评论不存在, id=%d", in.Id)
			return &v1.Response{
				Code:    404,
				Message: "评论不存在",
			}, nil
		}
		l.Errorf("DeleteComment err: 获取评论失败, id=%d, err=%v", in.Id, err)
		return internal("获取评论失败"), nil
	}

	// 超级管理员或者员工 或者 评论作者本人 可以删除
	if user.Role != constants.SUPERADMIN && user.Role != constants.STAFF && user.UID != comment.UserId {
		return bad("没有权限删除该评论"), nil
	}

	deletedAt := uint64(time.Now().Unix())
	err = l.CommentDao.UserSoftDelete(l.ctx, comment.Id, deletedAt)
	if err != nil {
		l.Errorf("DeleteComment err: 删除评论失败, id=%d, err=%v", in.Id, err)
		return bad("删除评论失败"), nil
	}

	commentRelationTask := &tasks.CommentRelationTask{
		Type:           tasks.CommentRelationDelete,
		ContentType:    comment.Type,
		ContentID:      comment.TargetId,
		UID:            user.UID,
		CommentID:      comment.Id,
		ParentID:       comment.ParentId,
		ReplyCommentID: comment.ReplyCommentId,
		ReplyUserID:    comment.ReplyUserId,
	}
	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, commentRelationTask)
	if err != nil {
		l.Errorf("DeleteComment err: 入列评论关系任务失败, task=%+v, err=%v", commentRelationTask, err)
		return internal("删除评论失败"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "删除评论成功",
	}, nil
}
