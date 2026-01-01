package logic

import (
	"context"
	"database/sql"
	"google.golang.org/protobuf/types/known/anypb"
	"time"

	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AddArticleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ArticleDao      model.ArticleModel
	ArticleImageDao model.ArticleImageModel
}

func NewAddArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddArticleLogic {
	return &AddArticleLogic{
		ctx:             ctx,
		svcCtx:          svcCtx,
		Logger:          logx.WithContext(ctx),
		ArticleDao:      model.NewArticleModel(svcCtx.Mysql),
		ArticleImageDao: model.NewArticleImageModel(svcCtx.Mysql),
	}
}

// AddArticle 添加文章
func (l *AddArticleLogic) AddArticle(in *v1.AddArticleRequest) (*v1.Response, error) {
	// 添加文章只有 超级管理员 和 员工 才能添加
	user := middleware.MustGetUser(l.ctx)
	if user.UID == InvalidUserID || (user.Role != SUPERADMIN && user.Role != STAFF) || user.Status != UserStatusNormal {
		l.Logger.Errorf("AddArticle err: 用户未登录或登录状态异常")

		return &v1.Response{
			Code:    400,
			Message: "用户未登录或登录状态异常",
		}, nil
	}

	// 检验请求体内容
	if in.Name == "" {
		l.Logger.Errorf("AddArticle err: 文章名称为空")

		return &v1.Response{
			Code:    400,
			Message: "文章名称为空",
		}, nil
	}
	if in.Content == "" {
		l.Logger.Errorf("AddArticle err: 文章内容为空")

		return &v1.Response{
			Code:    400,
			Message: "文章内容为空",
		}, nil
	}
	if in.Description == "" {
		l.Logger.Errorf("AddArticle err: 文章摘要为空")

		return &v1.Response{
			Code:    400,
			Message: "文章摘要为空",
		}, nil
	}
	if in.Cover == "" {
		l.Logger.Errorf("AddArticle err: 封面为空")

		return &v1.Response{
			Code:    400,
			Message: "封面为空",
		}, nil
	}
	if in.PublishedAt <= 0 {
		in.PublishedAt = time.Now().Unix()
	}
	if in.Tags == nil {
		in.Tags = []string{"默认标签"}
	}
	if len(in.Images) != 0 {
		for _, image := range in.Images {
			if image.Url == "" {
				l.Logger.Errorf("AddArticle err: 图片地址为空")

				return &v1.Response{
					Code:    400,
					Message: "图片地址为空",
				}, nil
			}
			if image.Sort < 0 {
				l.Logger.Errorf("AddArticle err: 图片排序为负数")

				return &v1.Response{
					Code:    400,
					Message: "图片排序为负数",
				}, nil
			}
			if _, ok := ArticleImageMap[image.Tp]; !ok {
				l.Logger.Errorf("AddArticle err: 图片类型错误")

				return &v1.Response{
					Code:    400,
					Message: "图片类型错误",
				}, nil
			}
		}
	}

	// 正式添加文章
	var article model.Article
	var images []*model.ArticleImage
	// 事务
	err := l.svcCtx.Mysql.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 添加文章

		// 1. 构造
		now := time.Now()
		article = model.Article{
			Name:         in.Name,
			Url:          in.Url,
			Description:  sql.NullString{String: in.Description, Valid: true},
			Cover:        in.Cover,
			Content:      sql.NullString{String: in.Content, Valid: true},
			Author:       in.Author,
			PublishedAt:  time.Unix(in.PublishedAt, 0),
			LikeCount:    0,
			ViewCount:    0,
			CollectCount: 0,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		result, err := l.ArticleDao.InsertWithSession(ctx, session, &article)
		if err != nil {
			l.Logger.Errorf("AddArticle err: %v", err)

			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			l.Logger.Errorf("AddArticle err: %v", err)
			return err
		}
		article.Id = uint64(id)

		// 添加文章图片

		// 2. 构造
		images = make([]*model.ArticleImage, len(in.Images))
		for i, image := range in.Images {
			images[i] = &model.ArticleImage{
				ArticleId: article.Id,
				Url:       image.Url,
				Sort:      image.Sort,
				CreatedAt: now,
				Type:      image.Tp,
			}
		}
		err = l.ArticleImageDao.InsertBatchWithSession(ctx, session, images)
		if err != nil {
			l.Logger.Errorf("AddArticle err: %v", err)
			return err
		}

		// 添加文章标签
		// todo 这里需要限制一下标签的数量

		// 1. 构造
		tags := make([]*model.ArticleTag, len(in.Tags))
		for i, tag := range in.Tags {
			tags[i] = &model.ArticleTag{
				ArticleId: article.Id,
				Tag:       tag,
			}
		}
		err = model.NewArticleTagModel(l.svcCtx.Mysql).InsertBatchWithSession(ctx, session, tags)
		if err != nil {
			l.Logger.Errorf("AddArticle err: %v", err)
			return err
		}
		return nil
	})

	if err != nil {
		l.Logger.Errorf("AddArticle err: %v", err)
		return &v1.Response{
			Code:    400,
			Message: "添加文章失败",
		}, nil
	}

	imageRes := make([]*v1.ArticleImage, len(images))
	for i, image := range images {
		imageRes[i] = &v1.ArticleImage{
			Url:  image.Url,
			Sort: image.Sort,
			Tp:   image.Type,
		}
	}

	// 构造返回内容
	res := &v1.AddArticleResponse{Article: &v1.Article{
		Id:           article.Id,
		Name:         article.Name,
		Tag:          in.Tags,
		Images:       imageRes,
		Url:          article.Url,
		Description:  article.Description.String,
		Cover:        article.Cover,
		Content:      article.Content.String,
		Author:       article.Author,
		PublishedAt:  article.PublishedAt.Unix(),
		CreatedAt:    article.CreatedAt.Unix(),
		UpdatedAt:    article.UpdatedAt.Unix(),
		LikeCount:    article.LikeCount,
		ViewCount:    article.ViewCount,
		CollectCount: article.CollectCount,
	}}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Logger.Errorf("AddArticle err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "转换类型失败",
		}, nil
	}

	return &v1.Response{
		Code:    200,
		Message: "添加文章成功",
		Data:    resAny,
	}, nil
}
