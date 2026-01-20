package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/luyb177/XiaoAnBackend/content/internal/logic"
	"github.com/luyb177/XiaoAnBackend/content/internal/repo/redisqueue"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"log"

	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pkg/article/convert"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
)

type ArticleRelationHandler struct {
	svcCtx        *svc.ServiceContext
	ArticleDao    model.ArticleModel
	ArticleTagDao model.ArticleTagModel
}

func NewArticleRelationHandler(svcCtx *svc.ServiceContext) *ArticleRelationHandler {
	return &ArticleRelationHandler{
		svcCtx:        svcCtx,
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

	log.Printf("processing article relation task: %+v", articleTask)

	switch articleTask.Type {
	case tasks.ArticleRelationAdd:
		fmt.Println(1)
		return h.handleAdd(ctx, &articleTask)
	case tasks.ArticleRelationModify:
		return h.handleModify(ctx, &articleTask)
	case tasks.ArticleRelationDelete:
		return h.handleDelete(ctx, &articleTask)
	default:
		log.Printf("unknown task type: %s", articleTask.Type)
		return nil
	}
}

func (h *ArticleRelationHandler) handleAdd(ctx context.Context, task *tasks.ArticleRelationTask) error {
	// 事务
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 插入新标签
		tagModels := convert.ArticleTagsFromStrings(task.ArticleID, task.Tags)
		err := h.ArticleTagDao.InsertBatchWithSession(ctx, session, tagModels)
		if err != nil {
			return err
		}
		// 2. 更新文章关联状态
		err = h.ArticleDao.UpdateRelationStatusWithSession(ctx, session, task.ArticleID, logic.RelationStatusNormal)
		if err != nil {
			return err
		}
		return nil
	})
}

func (h *ArticleRelationHandler) handleModify(ctx context.Context, task *tasks.ArticleRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 删除旧标签
		err := h.ArticleTagDao.DeleteBatchByArticleId(ctx, task.ArticleID)
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
	return h.ArticleTagDao.DeleteBatchByArticleId(ctx, task.ArticleID)
}
