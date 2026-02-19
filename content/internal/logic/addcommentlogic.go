package logic

import (
	"context"
	"time"

	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
)

type AddCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	CommentDao model.CommentModel
}

func NewAddCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddCommentLogic {
	return &AddCommentLogic{
		ctx:        ctx,
		svcCtx:     svcCtx,
		Logger:     logx.WithContext(ctx),
		CommentDao: model.NewCommentModel(svcCtx.Mysql),
	}
}

// AddComment 添加评论 最终一致性
func (l *AddCommentLogic) AddComment(in *v1.AddCommentRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或状态异常"), nil
	}

	ip, ok := middleware.GetIPInfo(l.ctx)
	if !ok {
		ip = middleware.DefaultIPInfo()
	}

	// 参数校验
	if resp := l.validate(in); resp != nil {
		l.Errorf("AddComment err: 参数校验失败, %s", resp.Message)
		return resp, nil
	}

	// TODO: user.Nickname user.Avatar 未来可以使用RPC调用服务来获取用户最新信息
	now := time.Now()
	comment := &model.Comment{
		Type:            in.Type,
		TargetId:        in.TargetId,
		UserId:          user.UID,
		Nickname:        in.Nickname,
		Avatar:          in.Avatar,
		IpLocation:      ip.City, // 使用 城市
		ParentId:        in.ParentId,
		ReplyCommentId:  in.ReplyCommentId,
		ReplyUserId:     in.ReplyUserId,
		Content:         in.Content,
		LikeCount:       0,
		SubCommentCount: 0,
		Status:          in.Status,
		CreatedAt:       now,
		UpdatedAt:       now,
		DeletedAt:       0,
		IsCounted:       0,
	}
	result, err := l.CommentDao.Insert(l.ctx, comment)
	if err != nil {
		l.Errorf("AddComment err: 添加评论失败, %v", err)
		return internal(err.Error()), nil
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		l.Errorf("AddComment err: 获取评论ID失败, %v", err)
		return internal("获取评论ID失败"), nil
	}
	comment.Id = uint64(commentID)

	commentRelationTask := &tasks.CommentRelationTask{
		Type:           tasks.CommentRelationAdd,
		ContentType:    in.Type,
		ContentID:      in.TargetId,
		UID:            user.UID,
		CommentID:      comment.Id,
		ParentID:       in.ParentId,
		ReplyCommentID: in.ReplyCommentId,
		ReplyUserID:    in.ReplyUserId,
	}

	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, commentRelationTask)
	if err != nil {
		l.Errorf("AddComment err: 入队列失败, %v", err)
		return internal("评论处理失败"), nil
	}

	res := &v1.AddCommentResponse{
		Id:             comment.Id,
		RelationStatus: RelationStatusPending,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("AddComment err: 组装返回结果失败, %v", err)
		return internal("组装返回结果失败"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "添加评论成功",
		Data:    resAny,
	}, nil
}

func (l *AddCommentLogic) validate(in *v1.AddCommentRequest) *v1.Response {
	switch {
	case in.Type == "":
		return bad("评论类型不能为空")
	case !isValidContentType(in.Type):
		return bad("评论类型不合法")
	case in.TargetId <= 0:
		return bad("评论目标ID必须大于0")
	case in.Nickname == "":
		return bad("评论昵称不能为空")
	case in.Avatar == "":
		return bad("评论头像不能为空")
	case in.Content == "":
		return bad("评论内容不能为空")
	case !isValidCommentStatus(in.Status):
		return bad("评论状态不合法")
	}
	return nil
}
