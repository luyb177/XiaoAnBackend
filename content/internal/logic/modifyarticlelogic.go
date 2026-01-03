package logic

import (
	"context"
	"database/sql"
	"errors"
	"github.com/luyb177/XiaoAnBackend/content/utils"
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
// 当然了，异步的话，还是应该使用异步队列的，这样之后可以修复
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

	go func(articleID uint64, tags []string, images []*v1.ArticleImage) {
		defer func() {
			if r := recover(); r != nil {
				l.Logger.Errorf("panic in async article update: %v", r)
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		const maxRetry = 3

		for attempt := 1; attempt <= maxRetry; attempt++ {
			err := l.svcCtx.Mysql.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
				// 1. 删除旧标签
				if err := l.ArticleTagDao.DeleteBatchByArticleIdWithSession(txCtx, session, articleID); err != nil {
					return err
				}

				// 2. 插入新标签
				tagModels := utils.ArticleTagsFromStrings(articleID, tags)
				if err := l.ArticleTagDao.InsertBatchWithSession(txCtx, session, tagModels); err != nil {
					return err
				}

				// 3. 删除旧图片
				if err := l.ArticleImageDao.DeleteBatchByArticleIdWithSession(txCtx, session, articleID); err != nil {
					return err
				}
				// 4. 插入新图片
				imageModels := utils.ArticleImagesFromPB(articleID, images)
				if err := l.ArticleImageDao.InsertBatchWithSession(txCtx, session, imageModels); err != nil {
					return err
				}

				// 5. 更新文章的 relation_status 为正常
				return l.ArticleDao.UpdateRelationStatusWithSession(txCtx, session, articleID, RelationStatusNormal)
			})

			if err == nil {
				// 成功执行，退出循环
				return
			}

			// 失败，记录日志并稍作延迟重试
			l.Logger.Errorf("async modify article attempt %d failed: %v", attempt, err)
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond) // 指数退避
		}

		l.Logger.Errorf("async modify article ultimately failed after %d attempts", maxRetry)

	}(in.Id, in.Tag, in.Images)

	// 构造返回值
	res := &v1.ModifyArticleResponse{
		Id:             in.Id,
		RelationStatus: RelationStatusPending,
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
