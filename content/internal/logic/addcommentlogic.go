package logic

import (
	"context"
	"fmt"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commentDao model.CommentModel
}

func NewAddCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddCommentLogic {
	return &AddCommentLogic{
		ctx:        ctx,
		svcCtx:     svcCtx,
		Logger:     logx.WithContext(ctx),
		commentDao: model.NewCommentModel(svcCtx.Mysql),
	}
}

// AddComment 添加评论
func (l *AddCommentLogic) AddComment(in *v1.AddCommentRequest) (*v1.Response, error) {
	userId := l.ctx.Value("user_id").(uint64)
	userRole := l.ctx.Value("user_role").(string)
	userStatus := l.ctx.Value("user_status").(int64)

	if userId == 0 || userRole == "" || userStatus != 1 {
		return &v1.Response{
			Code:    400,
			Message: "用户信息错误",
		}, fmt.Errorf("用户信息错误")
	}

	if in.Type == "" || in.TargetId <= 0 || in.Content == "" {
		return &v1.Response{
			Code:    400,
			Message: "参数错误",
		}, fmt.Errorf("参数错误")
	}

	if in.ParentId < 0 {
		in.ParentId = 0 // 根评论
	}
	if in.ReplyCommentId < 0 {
		in.ReplyCommentId = 0 // 没有@其他用户
	}

	now := time.Now()
	comment := model.Comment{
		Type:           in.Type,
		TargetId:       in.TargetId,
		UserId:         userId,
		ParentId:       in.ParentId,
		ReplyCommentId: in.ReplyCommentId,
		ReplyUserId:    in.ReplyUserId,
		Content:        in.Content,
		LikeCount:      0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	_, err := l.commentDao.Insert(l.ctx, &comment)
	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: "添加评论失败",
		}, fmt.Errorf("添加评论失败")
	}

	return &v1.Response{
		Code:    200,
		Message: "添加评论成功",
	}, nil
}
