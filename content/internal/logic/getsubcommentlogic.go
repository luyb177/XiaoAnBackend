package logic

import (
	"context"
	"errors"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/pkg/comment/convert"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSubCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	CommentDao model.CommentModel
}

func NewGetSubCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSubCommentLogic {
	return &GetSubCommentLogic{
		ctx:        ctx,
		svcCtx:     svcCtx,
		Logger:     logx.WithContext(ctx),
		CommentDao: model.NewCommentModel(svcCtx.Mysql),
	}
}

// GetSubComment 获取子评论
func (l *GetSubCommentLogic) GetSubComment(in *v1.GetSubCommentRequest) (*v1.Response, error) {
	if resp := l.validate(in); resp != nil {
		l.Errorf("GetSubComment err: 参数校验失败, %s", resp.Message)
		return resp, nil
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.PageSize <= 0 {
		in.PageSize = 10
	}

	offset := (in.Page - 1) * in.PageSize
	subCommentsModel, err := l.CommentDao.FindSubByTypeAndTargetIdAndParentId(l.ctx, in.ContentType, in.ContentId, in.ParentCommentId, offset, in.PageSize)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Errorf("GetSubComment err: 子评论不存在, contentType: %s, contentId: %d, parentCommentId: %d", in.ContentType, in.ContentId, in.ParentCommentId)
			return &v1.Response{
				Code:    404,
				Message: "没有更多子评论了",
			}, nil
		}
		l.Errorf("GetSubComment err: 获取子评论失败, %v", err)
		return internal("获取子评论失败"), nil
	}

	// todo 点赞
	subCommentsPB := convert.PBFromComment(subCommentsModel)

	res := &v1.GetSubCommentResponse{Comments: subCommentsPB}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("GetSubComment err: 响应封装失败, %v", err)
		return internal("响应封装失败"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "获取子评论成功",
		Data:    resAny,
	}, nil
}

func (l *GetSubCommentLogic) validate(in *v1.GetSubCommentRequest) *v1.Response {
	switch {
	case in.ContentType == "":
		return bad("内容类型不能为空")
	case in.ContentType != ContentTypeArticle && in.ContentType != ContentTypeComic && in.ContentType != ContentTypeVideo && in.ContentType != ContentTypePodcast:
		return bad("内容类型不合法")
	case in.ContentId <= 0:
		return bad("内容ID不合法")
	case in.ParentCommentId <= 0:
		return bad("父评论ID不合法")
	}
	return nil
}
