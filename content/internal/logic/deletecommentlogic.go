package logic

import (
	"context"
	"errors"
	"fmt"

	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commentDao model.CommentModel
}

func NewDeleteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCommentLogic {
	return &DeleteCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteComment 删除评论
func (l *DeleteCommentLogic) DeleteComment(in *v1.DeleteCommentRequest) (*v1.Response, error) {
	userId := l.ctx.Value("user_id").(uint64)
	userRole := l.ctx.Value("user_role").(string)
	userStatus := l.ctx.Value("user_status").(int64)

	if userId == 0 || userRole == "" || userStatus != 1 {
		return &v1.Response{
			Code:    400,
			Message: "用户信息错误",
		}, fmt.Errorf("用户信息错误")
	}

	if in.Id <= 0 {
		return &v1.Response{
			Code:    400,
			Message: "参数错误",
		}, fmt.Errorf("参数错误")
	}

	comment, err := l.commentDao.FindOne(l.ctx, in.Id)
	if !errors.Is(err, model.ErrNotFound) {
		return &v1.Response{
			Code:    400,
			Message: "评论不存在",
		}, fmt.Errorf("评论不存在")
	}
	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: "查询评论失败",
		}, fmt.Errorf("查询评论失败")
	}

	if comment.UserId != userId && userRole != "superadmin" && userRole != "admin" && userRole != "classadmin" && userRole != "teacher" && userRole != "student" {
		return &v1.Response{
			Code:    400,
			Message: "无权限删除该评论",
		}, fmt.Errorf("无权限删除该评论")
	}
	if err = l.commentDao.Delete(l.ctx, in.Id); err != nil {
		return &v1.Response{
			Code:    400,
			Message: "删除评论失败",
		}, fmt.Errorf("删除评论失败")
	}

	return &v1.Response{
		Code:    200,
		Message: "删除评论成功",
	}, nil
}
