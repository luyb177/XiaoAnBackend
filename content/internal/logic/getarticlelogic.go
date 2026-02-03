package logic

import (
	"context"
	"errors"

	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/article/convert"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"

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

// GetArticle 获取文章详细内容
func (l *GetArticleLogic) GetArticle(in *v1.GetArticleRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == InvalidUserID || user.Status != UserStatusNormal {
		return bad("用户未登录或登录状态异常"), nil
	}

	if in.Id == 0 {
		return bad("参数不合法"), nil
	}

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
			return internal("获取文章失败"), nil
		}
	}

	// 入队
	articleRelationTask := &tasks.ArticleRelationTask{
		Type:      tasks.ArticleRelationGet,
		ArticleID: in.Id,
		Tags:      nil,
	}

	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, articleRelationTask)
	if err != nil {
		l.Errorf("GetArticle err: 入队获取文章相关内容失败, %v", err)
		// 不影响获取文章内容
	}

	// 异步获取 tag like collect
	type TagResult struct {
		tags []*model.ArticleTag
		err  error
	}
	type LikeResult struct {
		liked bool
		err   error
	}
	type CollectResult struct {
		collected bool
		err       error
	}

	tagCh := make(chan TagResult, 1)
	likeCh := make(chan LikeResult, 1)
	collectCh := make(chan CollectResult, 1)

	go func() {
		t, err := l.ArticleTagDao.FindManyByArticleId(l.ctx, in.Id)
		tagCh <- TagResult{tags: t, err: err}
	}()

	go func() {
		liked, err := l.svcCtx.LikeRepo.HasLiked(l.ctx, user.UID, ContentTypeArticle, in.Id)
		likeCh <- LikeResult{liked: liked, err: err}
	}()

	go func() {
		collected, err := l.svcCtx.CollectRepo.HasCollect(l.ctx, user.UID, ContentTypeArticle, in.Id)
		collectCh <- CollectResult{collected: collected, err: err}
	}()

	// 等待 tag 结果
	tagsResult := <-tagCh

	if tagsResult.err != nil {
		l.Errorf("GetArticle err: %v", tagsResult.err)
		// 不影响获取文章内容
	}
	// 处理 tag
	tagsRes := convert.StringsFromArticleTags(tagsResult.tags)

	// 等待 like 结果
	likeRes := <-likeCh
	if likeRes.err != nil {
		l.Errorf("GetArticle err: %v", likeRes.err)
	}

	// 等待 collect 结果
	collectRes := <-collectCh
	if collectRes.err != nil {
		l.Errorf("GetArticle err: %v", collectRes.err)
	}

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
		IsLiked:        likeRes.liked,
		IsCollected:    collectRes.collected,
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
