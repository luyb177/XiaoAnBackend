package logic

import (
	"context"
	"database/sql"
	"errors"
	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteArticleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ArticleDao model.ArticleModel
}

func NewDeleteArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteArticleLogic {
	return &DeleteArticleLogic{
		ctx:        ctx,
		svcCtx:     svcCtx,
		Logger:     logx.WithContext(ctx),
		ArticleDao: model.NewArticleModel(svcCtx.Mysql),
	}
}

// DeleteArticle 删除文章
func (l *DeleteArticleLogic) DeleteArticle(in *v1.DeleteArticleRequest) (*v1.Response, error) {
	// 目前是只有超级管理员和员工可以删除文章
	user := middleware.MustGetUser(l.ctx)
	if user.UID == InvalidUserID || (user.Role != SUPERADMIN && user.Role != STAFF) || user.Status != UserStatusNormal {
		l.Errorf("DeleteArticle  err: 用户未登录或登录状态异常")

		return &v1.Response{
			Code:    400,
			Message: "用户未登录或登录状态异常",
		}, nil
	}

	// 请求参数验证
	validations := []Validation{
		{in.Id > 0, "文章ID参数错误"},
	}

	for _, v := range validations {
		if !v.Condition {
			l.Errorf("DeleteArticle err: %s", v.Message)

			return &v1.Response{
				Code:    400,
				Message: v.Message,
			}, nil
		}
	}

	// 查询有无
	article, err := l.ArticleDao.FindOneWithNotDelete(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &v1.Response{
				Code:    400,
				Message: "文章不存在",
			}, nil
		}
		return &v1.Response{
			Code:    400,
			Message: "查询文章时出错",
		}, nil
	}

	// 有 软删除
	deletedAt := uint64(time.Now().Unix())
	modifier := sql.NullInt64{Int64: int64(user.UID), Valid: true}
	err = l.ArticleDao.SoftDelete(l.ctx, article.Id, deletedAt, modifier)
	if err != nil {
		l.Errorf("DeleteArticle SoftDelete err: %v", err)
		return &v1.Response{
			Code:    500,
			Message: "删除文章失败",
		}, nil
	}

	// 将对应的 tag 全部删除
	articleRelationTask := &tasks.ArticleRelationTask{
		Type:      tasks.ArticleRelationDelete,
		ArticleID: article.Id,
		Tags:      nil,
	}
	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, articleRelationTask)
	if err != nil {
		l.Errorf("DeleteArticle Enqueue err: %v", err)
	}

	return &v1.Response{
		Code:    200,
		Message: "删除文章成功",
	}, nil
}
