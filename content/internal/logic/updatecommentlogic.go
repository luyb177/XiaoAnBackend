package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commentDao model.CommentModel
}

func NewUpdateCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCommentLogic {
	return &UpdateCommentLogic{
		ctx:        ctx,
		svcCtx:     svcCtx,
		Logger:     logx.WithContext(ctx),
		commentDao: model.NewCommentModel(svcCtx.Mysql),
	}
}

// UpdateComment 修改评论
func (l *UpdateCommentLogic) UpdateComment(in *v1.UpdateCommentRequest) (*v1.Response, error) {
	userId := l.ctx.Value("user_id").(uint64)
	userRole := l.ctx.Value("user_role").(string)
	userStatus := l.ctx.Value("user_status").(int64)

	if userId == 0 || userRole == "" || userStatus != 1 {
		return &v1.Response{
			Code:    400,
			Message: "用户信息错误",
		}, fmt.Errorf("用户信息错误")
	}

	if in.Id <= 0 || in.Content == "" {
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

	// 判断用户权限
	if comment.UserId != userId && userRole != SUPERADMIN && userRole != CLASSADMIN {
		return &v1.Response{
			Code:    400,
			Message: "用户权限不足",
		}, fmt.Errorf("用户权限不足")
	}

	comment.Content = in.Content
	comment.UpdatedAt = time.Now()

	err = l.commentDao.Update(l.ctx, comment)
	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: "更新评论失败",
		}, fmt.Errorf("更新评论失败")
	}

	return &v1.Response{
		Code:    200,
		Message: "更新评论成功",
	}, nil
}
