package logic

import (
	"context"
	"errors"

	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/comment/convert"
)

type GetRootCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	CommentDao model.CommentModel
}

func NewGetRootCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRootCommentLogic {
	return &GetRootCommentLogic{
		ctx:        ctx,
		svcCtx:     svcCtx,
		Logger:     logx.WithContext(ctx),
		CommentDao: model.NewCommentModel(svcCtx.Mysql),
	}
}

// GetRootComment 获取评论
func (l *GetRootCommentLogic) GetRootComment(in *v1.GetRootCommentRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或登录状态异常"), nil
	}

	if resp := l.validate(in); resp != nil {
		l.Errorf("GetRootComment err: 参数校验失败, %s", resp.Message)
		return resp, nil
	}

	if in.Page <= 0 {
		in.Page = 1
	}
	if in.PageSize <= 0 {
		in.PageSize = 10
	}

	offset := (in.Page - 1) * in.PageSize
	rootCommentsModel, err := l.CommentDao.FindRootByTypeAndTargetID(l.ctx, in.ContentType, in.ContentId, offset, in.PageSize)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &v1.Response{
				Code:    404,
				Message: "没有更多评论了",
			}, nil
		}
		l.Errorf("GetRootComment err: 获取根评论失败, type=%s, targetId=%d, err=%v", in.ContentType, in.ContentId, err)
		return internal("获取根评论失败"), nil
	}

	commentIDs := make([]uint64, len(rootCommentsModel))
	for i, comment := range rootCommentsModel {
		commentIDs[i] = comment.Id
	}

	type LikeResult struct {
		likes map[uint64]bool
		err   error
	}

	likeCh := make(chan LikeResult, 1)

	go func() {
		likes, err := l.svcCtx.LikeRepo.BatchHasLiked(l.ctx, user.UID, ContentTypeComment, commentIDs)
		likeCh <- LikeResult{likes: likes, err: err}
	}()

	// 等待结果
	likeResult := <-likeCh
	if likeResult.err != nil {
		l.Errorf("GetRootComment err: 获取点赞详情出错,%v", likeResult.err)
	}

	// todo 是否点赞
	rootCommentPB := convert.PBFromComment(rootCommentsModel, likeResult.likes)

	res := &v1.GetCommentResponse{Comments: rootCommentPB}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("GetRootComment err: 封装相应失败, type=%s, targetId=%d, err=%v", in.ContentType, in.ContentId, err)
		return internal("封装相应失败"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "获取根评论成功",
		Data:    resAny,
	}, nil
}

func (l *GetRootCommentLogic) validate(in *v1.GetRootCommentRequest) *v1.Response {
	switch {
	case in.ContentType == "":
		return bad("内容类型不能为空")
	case in.ContentType != ContentTypeArticle && in.ContentType != ContentTypeComic && in.ContentType != ContentTypePodcast && in.ContentType != ContentTypeVideo:
		return bad("内容类型不合法")
	case in.ContentId <= 0:
		return bad("内容ID为空或不合法")
	}
	return nil
}
