package logic

import (
	"context"
	"errors"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetArticleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ArticleDao      model.ArticleModel
	ArticleTagDao   model.ArticleTagModel
	ArticleImageDao model.ArticleImageModel
}

func NewGetArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetArticleLogic {
	return &GetArticleLogic{
		ctx:             ctx,
		svcCtx:          svcCtx,
		Logger:          logx.WithContext(ctx),
		ArticleDao:      model.NewArticleModel(svcCtx.Mysql),
		ArticleTagDao:   model.NewArticleTagModel(svcCtx.Mysql),
		ArticleImageDao: model.NewArticleImageModel(svcCtx.Mysql),
	}
}

// GetArticle 获取文章详细内容，无需登录
func (l *GetArticleLogic) GetArticle(in *v1.GetArticleRequest) (*v1.Response, error) {
	if in.Id <= 0 {
		l.Logger.Errorf("GetArticle err: 参数错误")

		return &v1.Response{
			Code:    400,
			Message: "参数错误",
		}, nil
	}

	// 获取文章
	article, err := l.ArticleDao.FindOneWithNotDelete(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("GetArticle err: 文章不存在")

			return &v1.Response{
				Code:    404,
				Message: "文章不存在",
			}, nil
		} else {
			l.Logger.Errorf("GetArticle err: %v", err)

			return &v1.Response{
				Code:    500,
				Message: "系统内部错误",
			}, nil
		}
	}

	// 异步获取 tag 和 image
	type tagResult struct {
		tags []*model.ArticleTag
		err  error
	}
	type imageResult struct {
		images []*model.ArticleImage
		err    error
	}

	tagCh := make(chan tagResult, 1)
	imageCh := make(chan imageResult, 1)

	go func() {
		t, err := l.ArticleTagDao.FindManyByArticleId(l.ctx, article.Id)
		tagCh <- tagResult{tags: t, err: err}
	}()

	go func() {
		i, err := l.ArticleImageDao.FindManyByArticleId(l.ctx, article.Id)
		imageCh <- imageResult{images: i, err: err}
	}()

	tagsResult := <-tagCh
	imagesResult := <-imageCh

	if tagsResult.err != nil {
		l.Logger.Errorf("GetArticle err: %v", tagsResult.err)
		// 不影响获取文章内容
	}
	if imagesResult.err != nil {
		l.Logger.Errorf("GetArticle err: %v", imagesResult.err)
		// 不影响获取文章内容
	}

	// 处理 tag
	var tagsRes []string
	for _, tag := range tagsResult.tags {
		tagsRes = append(tagsRes, tag.Tag)
	}

	// 处理 image
	var imagesRes []*v1.ArticleImage
	for _, image := range imagesResult.images {
		imagesRes = append(imagesRes, &v1.ArticleImage{
			Url:  image.Url,
			Sort: image.Sort,
			Tp:   image.Type,
		})
	}

	// 构造返回内容
	res := &v1.GetArticleResponse{Article: &v1.Article{
		Id:             article.Id,
		Name:           article.Name,
		Tag:            tagsRes,
		Images:         imagesRes,
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
	}}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Logger.Errorf("GetArticle err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "系统内部错误",
		}, nil
	}

	msg := "获取文章成功"
	if article.RelationStatus == RelationStatusPending {
		msg = "文章内容已更新，图片/标签同步中"
	}

	return &v1.Response{
		Code:    200,
		Message: msg,
		Data:    resAny,
	}, nil
}
