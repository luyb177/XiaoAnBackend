package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/luyb177/XiaoAnBackend/content/internal/logic"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/repo/redisqueue"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pkg/article/convert"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ArticleRelationHandler struct {
	logx.Logger
	svcCtx        *svc.ServiceContext
	ArticleDao    model.ArticleModel
	ArticleTagDao model.ArticleTagModel
}

func NewArticleRelationHandler(svcCtx *svc.ServiceContext, ctx context.Context) *ArticleRelationHandler {
	return &ArticleRelationHandler{
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		ArticleDao:    model.NewArticleModel(svcCtx.Mysql),
		ArticleTagDao: model.NewArticleTagModel(svcCtx.Mysql),
	}
}

func (h *ArticleRelationHandler) Handle(ctx context.Context, task taskqueue.Task) error {
	payload, err := task.Payload()
	if err != nil {
		return err
	}

	var rawTask redisqueue.RawTask
	err = json.Unmarshal(payload, &rawTask)
	if err != nil {
		return err
	}

	articleTask := tasks.ArticleRelationTask{}
	err = json.Unmarshal(rawTask.Data, &articleTask)
	if err != nil {
		return err
	}

	h.Infof("processing article relation task: %+v", articleTask)

	switch articleTask.Type {
	case tasks.ArticleRelationAdd:
		return h.handleAdd(ctx, &articleTask)
	case tasks.ArticleRelationModify:
		return h.handleModify(ctx, &articleTask)
	case tasks.ArticleRelationDelete:
		return h.handleDelete(ctx, &articleTask)
	default:
		h.Errorf("unknown task type: %s", articleTask.Type)
		return nil
	}
}

func (h *ArticleRelationHandler) handleAdd(ctx context.Context, task *tasks.ArticleRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 插入新标签
		tagModels := convert.ArticleTagsFromStrings(task.ArticleID, task.Tags)
		err := h.ArticleTagDao.InsertBatchWithSession(ctx, session, tagModels)
		if err != nil {
			return err
		}

		// 更新文章关联状态
		err = h.ArticleDao.UpdateRelationStatusWithSession(ctx, session, task.ArticleID, logic.RelationStatusNormal)
		if err != nil {
			return err
		}
		return nil
	})
}

func (h *ArticleRelationHandler) handleModify(ctx context.Context, task *tasks.ArticleRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 删除旧标签
		deletedAt := uint64(time.Now().Unix())
		modifier := sql.NullInt64{Int64: int64(deletedAt), Valid: true}
		err := h.ArticleTagDao.SoftDeleteByArticleIdWithSession(ctx, session, task.ArticleID, deletedAt, modifier)
		if err != nil {
			return err
		}

		// 2. 插入新标签
		tagModels := convert.ArticleTagsFromStrings(task.ArticleID, task.Tags)
		err = h.ArticleTagDao.InsertBatchWithSession(ctx, session, tagModels)
		if err != nil {
			return err
		}

		// 3. 更新文章关联状态
		return h.ArticleDao.UpdateRelationStatusWithSession(ctx, session, task.ArticleID, logic.RelationStatusNormal)
	})
}

func (h *ArticleRelationHandler) handleDelete(ctx context.Context, task *tasks.ArticleRelationTask) error {
	// 1. 删除标签
	deletedAt := uint64(time.Now().Unix())
	modifier := sql.NullInt64{Int64: int64(deletedAt), Valid: true}
	return h.ArticleTagDao.SoftDeleteByArticleId(ctx, task.ArticleID, deletedAt, modifier)
}
