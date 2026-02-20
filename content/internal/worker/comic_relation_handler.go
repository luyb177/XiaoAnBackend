package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/luyb177/XiaoAnBackend/content/internal/logic"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pkg/comic/convert"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
	"github.com/luyb177/XiaoAnBackend/infra/queue"
	"github.com/luyb177/XiaoAnBackend/infra/queue/redisqueue"
)

type ComicRelationHandler struct {
	logx.Logger
	svcCtx         *svc.ServiceContext
	ComicDao       model.ComicModel
	ComicTagDao    model.ComicTagModel
	CommentDao     model.CommentModel
	ContentLikeDao model.ContentLikeModel
}

func NewComicRelationHandler(ctx context.Context, svcCtx *svc.ServiceContext) *ComicRelationHandler {
	return &ComicRelationHandler{
		svcCtx:         svcCtx,
		Logger:         logx.WithContext(ctx),
		ComicDao:       model.NewComicModel(svcCtx.Mysql),
		ComicTagDao:    model.NewComicTagModel(svcCtx.Mysql),
		CommentDao:     model.NewCommentModel(svcCtx.Mysql),
		ContentLikeDao: model.NewContentLikeModel(svcCtx.Mysql),
	}
}

func (h *ComicRelationHandler) Handle(ctx context.Context, task queue.Task) error {
	payload, err := task.Payload()
	if err != nil {
		return err
	}

	var rawTask redisqueue.RawTask
	err = json.Unmarshal(payload, &rawTask)
	if err != nil {
		return err
	}

	comicTask := tasks.ComicRelationTask{}
	err = json.Unmarshal(rawTask.Data, &comicTask)
	if err != nil {
		return err
	}

	h.Infof("processing comic relation task: %+v", comicTask)

	switch comicTask.Type {
	case tasks.ComicRelationAdd:
		return h.handleAdd(ctx, &comicTask)
	case tasks.ComicRelationModify:
		return h.handleModify(ctx, &comicTask)
	case tasks.ComicRelationDelete:
		return h.handleDelete(ctx, &comicTask)
	case tasks.ComicRelationGet:
		return h.handleGet(ctx, &comicTask)
	default:
		h.Errorf("unknown comic relation task type: %v", comicTask.Type)
		return nil
	}
}

func (h *ComicRelationHandler) handleAdd(ctx context.Context, task *tasks.ComicRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 添加标签
		tagModels := convert.ComicTagsFromStrings(task.ComicID, task.Tags)
		err := h.ComicTagDao.InsertBatchWithSession(ctx, session, tagModels)
		if err != nil {
			return err
		}

		// 2. 更新关联状态
		return h.ComicDao.UpdateRelationStatusWithSession(ctx, session, task.ComicID, logic.RelationStatusNormal)
	})
}

func (h *ComicRelationHandler) handleModify(ctx context.Context, task *tasks.ComicRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 删除旧标签
		deletedAt := uint64(time.Now().Unix())
		err := h.ComicTagDao.SoftDeleteBatchByComicIDWithSession(ctx, session, task.ComicID, deletedAt)
		if err != nil {
			return err
		}

		// 2. 添加新标签
		tagModels := convert.ComicTagsFromStrings(task.ComicID, task.Tags)
		err = h.ComicTagDao.InsertBatchWithSession(ctx, session, tagModels)
		if err != nil {
			return err
		}

		// 3. 更新关联状态
		return h.ComicDao.UpdateRelationStatusWithSession(ctx, session, task.ComicID, logic.RelationStatusNormal)
	})
}

func (h *ComicRelationHandler) handleDelete(ctx context.Context, task *tasks.ComicRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 删除标签
		deletedAt := uint64(time.Now().Unix())
		err := h.ComicTagDao.SoftDeleteBatchByComicIDWithSession(ctx, session, task.ComicID, deletedAt)
		if err != nil {
			return err
		}

		// 2. 删除评论
		_, err = h.CommentDao.SoftDeleteByTypeAndTargetIDWithSession(ctx, session, logic.ContentTypeComic, task.ComicID, deletedAt)
		if err != nil {
			return err
		}

		// 删除点赞
		_, err = h.ContentLikeDao.SoftDeleteByTypeTargetIDWithSession(ctx, session, logic.ContentTypeComic, task.ComicID, deletedAt)
		if err != nil {
			return err
		}

		// 3. 删除对应的全部章节
		comicChapterRelationTask := &tasks.ComicChapterRelationTask{
			Type:    tasks.ComicChapterRelationDeleteAll,
			ComicID: task.ComicID,
			UID:     task.UID,
		}

		return h.svcCtx.TaskQueue.Enqueue(ctx, comicChapterRelationTask)
	})
}

func (h *ComicRelationHandler) handleGet(ctx context.Context, task *tasks.ComicRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 增加 漫画浏览量
		_, err := h.ComicDao.IncrViewCountWithSession(ctx, session, task.ComicID)
		return err
	})
}
