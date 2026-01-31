package logic

import (
	"context"

	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommentLogic {
	return &GetCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetComment 获取评论
func (l *GetCommentLogic) GetComment(in *v1.GetCommentRequest) (*v1.Response, error) {
	if resp := l.validate(in); resp != nil {
		l.Errorf("GetComment err: 参数校验失败, %s", resp.Message)
		return resp, nil
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.PageSize <= 0 {
		in.PageSize = 10
	}

	return &v1.Response{}, nil
}

func (l *GetCommentLogic) validate(in *v1.GetCommentRequest) *v1.Response {
	switch {
	case in.ContentId <= 0:
		return bad("内容ID无效")
	case in.ContentType == "":
		return bad("内容类型不能为空")
	case in.ContentType != ContentTypePodcast && in.ContentType != ContentTypeArticle && in.ContentType != ContentTypeVideo && in.ContentType != ContentTypeComic:
		return bad("内容类型无效")
	}
	return nil
}
