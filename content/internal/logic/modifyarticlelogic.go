package logic

import (
	"context"
	"database/sql"
	"errors"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/protobuf/types/known/anypb"
	"time"

	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyArticleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ArticleDao      model.ArticleModel
	ArticleTagDao   model.ArticleTagModel
	ArticleImageDao model.ArticleImageModel
}

func NewModifyArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyArticleLogic {
	return &ModifyArticleLogic{
		ctx:             ctx,
		svcCtx:          svcCtx,
		Logger:          logx.WithContext(ctx),
		ArticleDao:      model.NewArticleModel(svcCtx.Mysql),
		ArticleTagDao:   model.NewArticleTagModel(svcCtx.Mysql),
		ArticleImageDao: model.NewArticleImageModel(svcCtx.Mysql),
	}
}

// ModifyArticle 修改文章
// todo 存储修改历史
// 因为接口比较慢，所以事务中主要新增文章，图片和标签在事务外异步执行
// 放在异步中的话，需要一个状态来标记异步完成
func (l *ModifyArticleLogic) ModifyArticle(in *v1.ModifyArticleRequest) (*v1.Response, error) {
	user := middleware.MustGetUser(l.ctx)
	if user.UID == InvalidUserID || (user.Role != SUPERADMIN && user.Role != STAFF) || user.Status != UserStatusNormal {
		l.Logger.Errorf("ModifyArticle err: 用户登录状态异常")

		return &v1.Response{
			Code:    400,
			Message: "用户未登录或者登录状态异常",
		}, nil
	}

	// 验证参数
	if in.Id <= 0 {
		l.Logger.Errorf("ModifyArticle err: 文章ID不能小于等于0")

		return &v1.Response{
			Code:    400,
			Message: "文章ID不能小于等于0",
		}, nil
	}
	if in.Name == "" {
		l.Logger.Errorf("ModifyArticle err: 文章名称为空")

		return &v1.Response{
			Code:    400,
			Message: "文章名称为空",
		}, nil
	}
	if in.Author == "" {
		l.Logger.Errorf("ModifyArticle err: 文章作者为空")

		return &v1.Response{
			Code:    400,
			Message: "文章作者为空",
		}, nil
	}
	if in.Content == "" {
		l.Logger.Errorf("ModifyArticle err: 文章内容为空")

		return &v1.Response{
			Code:    400,
			Message: "文章内容为空",
		}, nil
	}
	if in.Description == "" {
		l.Logger.Errorf("ModifyArticle err: 文章摘要为空")

		return &v1.Response{
			Code:    400,
			Message: "文章摘要为空",
		}, nil
	}
	if len(in.Tag) == 0 {
		in.Tag = []string{"默认标签"}
	}
	if len(in.Tag) > 10 {
		l.Logger.Errorf("ModifyArticle err: 标签数量不能超过10")

		return &v1.Response{
			Code:    400,
			Message: "标签数量不能超过10",
		}, nil
	}
	// 检查标签
	for _, tag := range in.Tag {
		if tag == "" {
			l.Logger.Errorf("ModifyArticle err: 标签不能为空")

			return &v1.Response{
				Code:    400,
				Message: "标签不能为空",
			}, nil
		}
	}
	if len(in.Images) != 0 {
		for _, image := range in.Images {
			if image.Url == "" {
				l.Logger.Errorf("ModifyArticle err: 图片地址为空")

				return &v1.Response{
					Code:    400,
					Message: "图片地址为空",
				}, nil
			}
			if image.Sort < 0 {
				l.Logger.Errorf("ModifyArticle err: 图片排序为负数")

				return &v1.Response{
					Code:    400,
					Message: "图片排序为负数",
				}, nil
			}
			if _, ok := ArticleImageMap[image.Tp]; !ok {
				l.Logger.Errorf("ModifyArticle err: 图片类型错误")

				return &v1.Response{
					Code:    400,
					Message: "图片类型错误",
				}, nil
			}
		}
	}
	if in.PublishedAt <= 0 {
		in.PublishedAt = time.Now().Unix()
	}

	// 先验证文章是否存在或者是否被删除
	queryCtx, cancel := context.WithTimeout(l.ctx, 2*time.Second)
	defer cancel()

	article, err := l.ArticleDao.FindOneWithNotDelete(queryCtx, in.Id)
	switch {
	case errors.Is(err, model.ErrNotFound):
		l.Logger.Errorf("ModifyArticle err: 文章不存在")
		return &v1.Response{
			Code:    400,
			Message: "文章不存在",
		}, nil
	case err != nil:
		l.Logger.Errorf("ModifyArticle err: %v", err)
		return &v1.Response{
			Code:    400,
			Message: "修改文章错误",
		}, nil
	}

	txCtx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	// 构造修改内容 事务
	err = l.svcCtx.Mysql.TransactCtx(txCtx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 更新文章
		// 1.1 构造
		article.Name = in.Name
		article.Url = in.Url
		article.Description = sql.NullString{String: in.Description, Valid: true}
		article.Cover = in.Cover
		article.Content = sql.NullString{String: in.Content, Valid: true}
		article.Author = in.Author
		article.PublishedAt = time.Unix(in.PublishedAt, 0)
		article.LastModifiedBy = sql.NullInt64{Int64: int64(user.UID), Valid: true}
		// 标记待同步
		article.RelationStatus = RelationStatusPending

		// 1.2 更新
		return l.ArticleDao.UpdateWithSession(ctx, session, article)
	})

	if err != nil {
		l.Logger.Errorf("ModifyArticle err: %v", err)

		return &v1.Response{
			Code:    400,
			Message: "修改文章失败",
		}, nil
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tags := make([]*model.ArticleTag, len(in.Tag))
		for i, tag := range in.Tag {
			tags[i] = &model.ArticleTag{
				ArticleId: in.Id,
				Tag:       tag,
			}
		}
		images := make([]*model.ArticleImage, len(in.Images))
		for i, image := range in.Images {
			images[i] = &model.ArticleImage{
				ArticleId: in.Id,
				Url:       image.Url,
				Sort:      image.Sort,
				Type:      image.Tp,
			}
		}
		if err := l.ArticleTagDao.DeleteBatchByArticleId(ctx, in.Id); err != nil {
			l.Logger.Errorf("async delete article tag err: %v", err)
			return
		}
		if err := l.ArticleTagDao.InsertBatch(ctx, tags); err != nil {
			l.Logger.Errorf("async insert article tag err: %v", err)
			return
		}
		if err := l.ArticleImageDao.DeleteBatchByArticleId(ctx, in.Id); err != nil {
			l.Logger.Errorf("async delete article image err: %v", err)
			return
		}
		if err := l.ArticleImageDao.InsertBatch(ctx, images); err != nil {
			l.Logger.Errorf("async insert article image err: %v", err)
			return
		}

		// 将关系改为正常
		if err := l.ArticleDao.UpdateRelationStatus(ctx, in.Id, RelationStatusNormal); err != nil {
			l.Logger.Errorf("update relation_status err: %v", err)
		}
	}()

	// 构造返回值
	res := &v1.ModifyArticleResponse{
		Article: &v1.Article{
			Id:           in.Id,
			Name:         in.Name,
			Tag:          in.Tag,
			Images:       in.Images,
			Url:          in.Url,
			Description:  in.Description,
			Cover:        in.Cover,
			Content:      in.Content,
			Author:       in.Author,
			PublishedAt:  in.PublishedAt,
			CreatedAt:    article.CreatedAt.Unix(),
			UpdatedAt:    time.Now().Unix(),
			LikeCount:    article.LikeCount,
			ViewCount:    article.ViewCount,
			CollectCount: article.CollectCount,
		},
	}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Logger.Errorf("ModifyArticle err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "修改文章失败",
		}, nil
	}

	return &v1.Response{
		Code:    200,
		Message: "修改文章成功",
		Data:    resAny,
	}, nil
}
