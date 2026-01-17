package logic

import (
	"context"
	"fmt"
	"sync"

	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"
)

type GetCommentsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commentDao     model.CommentModel
	contentLikeDao model.ContentLikeModel
}

func NewGetCommentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommentsLogic {
	return &GetCommentsLogic{
		ctx:            ctx,
		svcCtx:         svcCtx,
		Logger:         logx.WithContext(ctx),
		commentDao:     model.NewCommentModel(svcCtx.Mysql),
		contentLikeDao: model.NewContentLikeModel(svcCtx.Mysql),
	}
}

// GetComments 获取根评论
// todo 限制并发量
// todo 减少DB请求 - 将跟评论的 id 使用 in(?,?) 来一次请求是否点赞
func (l *GetCommentsLogic) GetComments(in *v1.GetCommentsRequest) (*v1.Response, error) {
	userId := l.ctx.Value("user_id").(uint64)
	userRole := l.ctx.Value("user_role").(string)
	userStatus := l.ctx.Value("user_status").(int64)

	if userId == 0 || userRole == "" || userStatus != 1 {
		return &v1.Response{
			Code:    400,
			Message: "用户信息错误",
		}, fmt.Errorf("用户信息错误")
	}

	if in.Type == "" || in.TargetId <= 0 {
		return &v1.Response{
			Code:    400,
			Message: "参数错误",
		}, fmt.Errorf("参数错误")
	}

	if in.Page <= 0 {
		in.Page = 1
	}
	if in.PageSize <= 0 {
		in.PageSize = 10
	}

	offset := (in.Page - 1) * in.PageSize

	var wg sync.WaitGroup
	var comments []*model.Comment
	var total int64
	var commentsErr, countErr error

	// 异步获取根评论数据
	wg.Add(2)
	go func() {
		var err error
		comments, err = l.commentDao.FindByTypeAndTargetId(l.ctx, in.Type, in.TargetId, offset, in.PageSize)
		commentsErr = err
		wg.Done()
	}()
	go func() {
		var err error
		total, err = l.commentDao.CountParentByTypeAndTargetId(l.ctx, in.Type, in.TargetId)
		countErr = err
		wg.Done()
	}()
	wg.Wait()

	if commentsErr != nil || countErr != nil {
		return &v1.Response{Code: 400, Message: "获取评论数据失败"}, fmt.Errorf("commentsErr=%v countErr=%v", commentsErr, countErr)
	}

	commentRes := make([]*v1.CommentItem, 0, len(comments))
	for _, comment := range comments {
		// 先放一下跟评论的数据
		commentRes = append(commentRes, &v1.CommentItem{
			Comment: &v1.CommentDetail{
				Id:             comment.Id,
				Type:           comment.Type,
				TargetId:       comment.TargetId,
				UserId:         comment.UserId,
				ParentId:       comment.ParentId,
				ReplyCommentId: comment.ReplyCommentId,
				ReplyUserId:    comment.ReplyUserId,
				Content:        comment.Content,
				LikeCount:      comment.LikeCount,
				CreatedAt:      comment.CreatedAt.Unix(),
				UpdatedAt:      comment.UpdatedAt.Unix(),
				IsLiked:        false, // 这里需要查看
			},
			ChildPreview: nil,
			ChildTotal:   0,
		})
	}

	// 构造一下
	res := &v1.GetCommentsResponse{
		Comments: commentRes,
		Total:    total,
	}

	// 异步获取子评论数据
	for i := range res.Comments {
		c := res.Comments[i] // 闭包不要 capture 迭代变量，要 capture value 的副本
		wg.Add(1)

		go func(item *v1.CommentItem) {
			defer wg.Done()

			// 获取跟评论是否点赞
			//contentLike, err := l.contentLikeDao.FindOneByUserIdTypeTargetId(l.ctx, userId, TypeComment, item.Comment.Id)
			//if err == nil {
			//	// 有点赞数据，判断是否有效
			//	if contentLike.Status == Valid {
			//		item.Comment.IsLiked = true
			//	}
			//} else if !errors.Is(err, model.ErrNotFound) {
			//	l.Logger.Errorf("GetComments 获取跟评论点赞数据失败 err: %v，其中parent_id为 %d", err, item.Comment.Id)
			//}

			// 获取子评论总数
			var childTotal int64
			var childErr error
			childTotal, childErr = l.commentDao.CountChildByTypeAndTargetId(l.ctx, in.Type, in.TargetId, item.Comment.Id)
			if childErr != nil {
				l.Logger.Errorf("GetComments 获取子评论总数失败 err: %v，其中parent_id为 %d", childErr, item.Comment.Id)
				return
			}

			item.ChildTotal = childTotal
			// 无子评论
			if childTotal == 0 {
				return
			}

			// 获取子评论数据
			var children []*model.Comment
			children, childErr = l.commentDao.FindDefaultChildByTypeAndTargetId(l.ctx, in.Type, in.TargetId, item.Comment.Id)
			if childErr == nil {
				// 构造子评论数据
				childComments := make([]*v1.CommentDetail, 0, len(children))
				for _, child := range children {
					childComments = append(childComments, &v1.CommentDetail{
						Id:             child.Id,
						Type:           child.Type,
						TargetId:       child.TargetId,
						UserId:         child.UserId,
						ParentId:       child.ParentId,
						ReplyCommentId: child.ReplyCommentId,
						ReplyUserId:    child.ReplyUserId,
						Content:        child.Content,
						LikeCount:      child.LikeCount,
						CreatedAt:      child.CreatedAt.Unix(),
						UpdatedAt:      child.UpdatedAt.Unix(),
						IsLiked:        false,
					})
				}
				item.ChildPreview = childComments

				//	获取子评论是否点赞
				//for _, cp := range item.ChildPreview {
				//	childContentLike, err := l.contentLikeDao.FindOneByUserIdTypeTargetId(l.ctx, userId, TypeComment, cp.Id)
				//	if err == nil {
				//		if childContentLike.Status == Valid {
				//			cp.IsLiked = true
				//		}
				//	} else if !errors.Is(err, model.ErrNotFound) {
				//		l.Logger.Errorf("GetComments 获取子评论点赞数据失败 err: %v，其中parent_id为 %d", err, cp.Id)
				//	}
				//
				//	// 有点赞数据，判断是否有效
				//	if childContentLike.Status == Valid {
				//		cp.IsLiked = true
				//	}
				//}

			} else {
				l.Logger.Errorf("GetComments 获取子评论数据失败 err: %v，其中parent_id为 %d", childErr, item.Comment.Id)
			}
		}(c)
	}
	wg.Wait()

	resAny, err := anypb.New(res)
	if err != nil {
		return &v1.Response{
			Code:    400,
			Message: "消息类型转换失败",
		}, err
	}
	return &v1.Response{
		Code:    200,
		Message: "获取评论数据成功",
		Data:    resAny,
	}, nil
}
