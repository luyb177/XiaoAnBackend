package logic

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
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
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == InvalidUserID || (user.Role != SUPERADMIN && user.Role != STAFF) || user.Status != UserStatusNormal {
		return bad("用户未登录或状态异常"), nil
	}
	validations := []Validation{
		{in.Name != "", "文章名称不能为空"},
		{in.Content != "", "文章内容不能为空"},
		{in.Description != "", "文章摘要不能为空"},
		{in.Cover != "", "封面不能为空"},
		{len(in.Tags) <= 10, "标签数量超出限制"},
	}
	for _, v := range validations {
		if !v.Condition {
			l.Errorf("AddArticle err: %s", v.Message)
			return &v1.Response{
				Code:    400,
				Message: v.Message,
			}, nil
		}
	}

	// 设置默认值
	now := time.Now()
	if in.PublishedAt <= 0 {
		in.PublishedAt = now.Unix()
	}
	if len(in.Tags) == 0 {
		in.Tags = []string{"默认标签"}
	}

	// 添加文章
	article := model.Article{
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
		CommentCount:   0,
		DeletedAt:      0,
	}

	// 写入
	result, err := l.ArticleDao.Insert(l.ctx, &article)
	if err != nil {
		l.Errorf("insert article error: %v", err)
		return &v1.Response{
			Code:    400,
			Message: "添加文章失败",
		}, nil
	}

	// 回写
	id, err := result.LastInsertId()
	if err != nil {
		l.Errorf("get last insert id error: %v", err)
		return &v1.Response{
			Code:    400,
			Message: "添加文章失败",
		}, nil
	}
	article.Id = uint64(id)

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
