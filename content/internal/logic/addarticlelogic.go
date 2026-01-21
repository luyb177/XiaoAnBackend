package logic

import (
	"context"
	"database/sql"
	"time"

	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/protobuf/types/known/anypb"
)

type AddArticleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ArticleDao model.ArticleModel
}

func NewAddArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddArticleLogic {
	return &AddArticleLogic{
		ctx:        ctx,
		svcCtx:     svcCtx,
		Logger:     logx.WithContext(ctx),
		ArticleDao: model.NewArticleModel(svcCtx.Mysql),
	}
}

// AddArticle 添加文章
func (l *AddArticleLogic) AddArticle(in *v1.AddArticleRequest) (*v1.Response, error) {
	// 添加文章只有 超级管理员 和 员工 才能添加
	user := middleware.MustGetUser(l.ctx)
	if user.UID == InvalidUserID || (user.Role != SUPERADMIN && user.Role != STAFF) || user.Status != UserStatusNormal {
		l.Errorf("AddArticle err: 用户未登录或无权限")

		return &v1.Response{
			Code:    400,
			Message: "用户未登录或无权限",
		}, nil
	}

	// 检验请求体内容
	if in.Name == "" {
		l.Errorf("AddArticle err: 文章名称为空")

		return &v1.Response{
			Code:    400,
			Message: "文章名称为空",
		}, nil
	}
	if in.Content == "" {
		l.Errorf("AddArticle err: 文章内容为空")

		return &v1.Response{
			Code:    400,
			Message: "文章内容为空",
		}, nil
	}
	if in.Description == "" {
		l.Errorf("AddArticle err: 文章摘要为空")

		return &v1.Response{
			Code:    400,
			Message: "文章摘要为空",
		}, nil
	}
	if in.Cover == "" {
		l.Errorf("AddArticle err: 封面为空")

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
	if len(in.Tags) > 10 {
		l.Errorf("AddArticle err: 标签数量超出限制")

		return &v1.Response{
			Code:    400,
			Message: "标签数量超出限制",
		}, nil
	}

	// 正式添加文章
	// 事务
	// todo 只添加一个不需要事务，暂时先不改

	var article model.Article
	now := time.Now()
	err := l.svcCtx.Mysql.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 添加文章
		// 1. 构造
		article = model.Article{
			Name:           in.Name,
			Url:            in.Url,
			Description:    sql.NullString{String: in.Description, Valid: true},
			Cover:          in.Cover,
			Content:        sql.NullString{String: in.Content, Valid: true},
			Author:         in.Author,
			PublishedAt:    time.Unix(in.PublishedAt, 0),
			RelationStatus: RelationStatusPending,
			LastModifiedBy: sql.NullInt64{Int64: int64(user.UID), Valid: true},
			LikeCount:      0,
			ViewCount:      0,
			CollectCount:   0,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		result, err := l.ArticleDao.InsertWithSession(ctx, session, &article)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		article.Id = uint64(id)

		return nil
	})

	if err != nil {
		l.Errorf("AddArticle err: %v", err)

		return &v1.Response{
			Code:    400,
			Message: "添加文章失败",
		}, nil
	}

	articleRelationTask := &tasks.ArticleRelationTask{
		Type:      tasks.ArticleRelationAdd,
		ArticleID: article.Id,
		Tags:      in.Tags,
	}

	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, articleRelationTask)
	if err != nil {
		l.Errorf("AddArticle Enqueue err: %v", err)
	}

	// 构造返回内容
	res := &v1.AddArticleResponse{
		Id:             article.Id,
		RelationStatus: RelationStatusPending,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("AddArticle err: %v", err)

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
