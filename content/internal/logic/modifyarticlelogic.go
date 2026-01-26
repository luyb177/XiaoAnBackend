package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"
)

type ModifyArticleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ArticleDao    model.ArticleModel
	ArticleTagDao model.ArticleTagModel
}

func NewModifyArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyArticleLogic {
	return &ModifyArticleLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		ArticleDao:    model.NewArticleModel(svcCtx.Mysql),
		ArticleTagDao: model.NewArticleTagModel(svcCtx.Mysql),
	}
}

// ModifyArticle 修改文章
// todo: 修改历史
func (l *ModifyArticleLogic) ModifyArticle(in *v1.ModifyArticleRequest) (*v1.Response, error) {
	user := middleware.MustGetUser(l.ctx)
	if user.UID == InvalidUserID || (user.Role != SUPERADMIN && user.Role != STAFF) || user.Status != UserStatusNormal {
		l.Logger.Errorf("ModifyArticle err: 用户未登录或者没有权限")

		return &v1.Response{
			Code:    400,
			Message: "用户未登录或者没有权限",
		}, nil
	}

	// 验证参数
	validations := []Validation{
		{in.Id > 0, "文章ID不能小于等于0"},
		{in.Name != "", "文章名称为空"},
		{in.Author != "", "文章作者为空"},
		{in.Content != "", "文章内容为空"},
		{in.Description != "", "文章摘要为空"},
		{len(in.Tag) <= 10, "标签数量不能超过10"},
	}

	for _, v := range validations {
		if !v.Condition {
			l.Errorf("ModifyArticle err: %s", v.Message)

			return &v1.Response{
				Code:    400,
				Message: v.Message,
			}, nil
		}
	}

	if in.Tag == nil || len(in.Tag) == 0 {
		in.Tag = []string{"默认标签"}
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
	if in.PublishedAt <= 0 {
		in.PublishedAt = time.Now().Unix()
	}

	// 查询文章是否存在
	article, err := l.ArticleDao.FindOneWithNotDelete(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("ModifyArticle err: 文章不存在")

			return &v1.Response{
				Code:    400,
				Message: "文章不存在",
			}, nil
		}
	}

	// 更新文章
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
	err = l.ArticleDao.Update(l.ctx, article)
	if err != nil {
		l.Logger.Errorf("ModifyArticle err: %v", err)

		return &v1.Response{
			Code:    400,
			Message: "修改文章失败",
		}, nil
	}

	articleRelationTask := &tasks.ArticleRelationTask{
		Type:      tasks.ArticleRelationModify,
		ArticleID: article.Id,
		UID:       user.UID,
		Tags:      in.Tag,
	}

	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, articleRelationTask)
	if err != nil {
		l.Logger.Errorf("ModifyArticle Enqueue err: %v", err)
	}

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
