package logic

import (
	"context"
	"errors"

	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/article/convert"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"
)

type GetArticleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ArticleDao    model.ArticleModel
	ArticleTagDao model.ArticleTagModel
}

func NewGetArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetArticleLogic {
	return &GetArticleLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		ArticleDao:    model.NewArticleModel(svcCtx.Mysql),
		ArticleTagDao: model.NewArticleTagModel(svcCtx.Mysql),
	}
}

// GetArticle 获取文章详细内容，无需登录
func (l *GetArticleLogic) GetArticle(in *v1.GetArticleRequest) (*v1.Response, error) {
	validations := []Validation{
		{in.Id > 0, "文章ID参数错误"},
	}

	for _, v := range validations {
		if !v.Condition {
			l.Errorf("GetArticle err: %s", v.Message)

			return &v1.Response{
				Code:    400,
				Message: v.Message,
			}, nil
		}
	}

	// 异步获取 tag
	type tagResult struct {
		tags []*model.ArticleTag
		err  error
	}

	tagCh := make(chan tagResult, 1)

	go func() {
		t, err := l.ArticleTagDao.FindManyByArticleId(l.ctx, in.Id)
		tagCh <- tagResult{tags: t, err: err}
	}()

	// 获取文章
	article, err := l.ArticleDao.FindOneWithNotDelete(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Errorf("GetArticle err: 文章不存在")

			return &v1.Response{
				Code:    404,
				Message: "文章不存在",
			}, nil
		} else {
			l.Errorf("GetArticle err: %v", err)

			return &v1.Response{
				Code:    500,
				Message: "系统内部错误",
			}, nil
		}
	}

	// 等待 tag 结果
	tagsResult := <-tagCh

	if tagsResult.err != nil {
		l.Errorf("GetArticle err: %v", tagsResult.err)
		// 不影响获取文章内容
	}
	// 处理 tag
	tagsRes := convert.StringsFromArticleTags(tagsResult.tags)

	// 构造返回内容
	res := &v1.GetArticleResponse{Article: &v1.Article{
		Id:             article.Id,
		Name:           article.Name,
		Tag:            tagsRes,
		Url:            article.Url,
		Description:    article.Description.String,
		Cover:          article.Cover,
		Content:        article.Content.String,
		Author:         article.Author,
		PublishedAt:    article.PublishedAt.Unix(),
		CreatedAt:      article.CreatedAt.Unix(),
		UpdatedAt:      article.UpdatedAt.Unix(),
		LikeCount:      article.LikeCount,
		ViewCount:      article.ViewCount,
		CollectCount:   article.CollectCount,
		LastModifiedBy: article.LastModifiedBy.Int64,
		RelationStatus: article.RelationStatus,
		CommentCount:   article.CommentCount,
	}}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("GetArticle err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "系统内部错误",
		}, nil
	}

	msg := "获取文章成功"
	if article.RelationStatus == RelationStatusPending {
		msg = "文章相关内容同步中,请稍后刷新查看"
	}

	return &v1.Response{
		Code:    200,
		Message: msg,
		Data:    resAny,
	}, nil
}
