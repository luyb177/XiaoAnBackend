package worker

import (
	"context"
	"encoding/json"
	"github.com/luyb177/XiaoAnBackend/content/internal/logic"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/repo/redisqueue"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pkg/comic/convert"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ComicChapterRelationHandler struct {
	logx.Logger
	svcCtx          *svc.ServiceContext
	ComicDao        model.ComicModel
	ComicChapterDao model.ComicChapterModel
	ComicPageDao    model.ComicPageModel
}

func NewComicChapterRelationHandler(svcCtx *svc.ServiceContext, ctx context.Context) *ComicChapterRelationHandler {
	return &ComicChapterRelationHandler{
		svcCtx:          svcCtx,
		Logger:          logx.WithContext(ctx),
		ComicChapterDao: model.NewComicChapterModel(svcCtx.Mysql),
		ComicPageDao:    model.NewComicPageModel(svcCtx.Mysql),
		ComicDao:        model.NewComicModel(svcCtx.Mysql),
	}
}

func (h *ComicChapterRelationHandler) Handle(ctx context.Context, task taskqueue.Task) error {
	payload, err := task.Payload()
	if err != nil {
		return err
	}

	var rawTask redisqueue.RawTask
	err = json.Unmarshal(payload, &rawTask)
	if err != nil {
		return err
	}

	comicChapterTask := tasks.ComicChapterRelationTask{}
	err = json.Unmarshal(rawTask.Data, &comicChapterTask)
	if err != nil {
		return err
	}

	h.Infof("processing comic chapter relation task: %+v", comicChapterTask)

	switch comicChapterTask.Type {
	case tasks.ComicChapterRelationAdd:
		return h.handleAdd(ctx, &comicChapterTask)
	case tasks.ComicChapterRelationModify:
		return h.handleModify(ctx, &comicChapterTask)
	case tasks.ComicChapterRelationDelete:
		return h.handleDelete(ctx, &comicChapterTask)
	case tasks.ComicChapterRelationDeleteAll:
		return h.handleDeleteAll(ctx, &comicChapterTask)
	default:
		h.Errorf("unknown comic chapter relation task type: %v", comicChapterTask.Type)
		return nil
	}
}
func (h *ComicChapterRelationHandler) handleAdd(ctx context.Context, task *tasks.ComicChapterRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 章节数+1
		err := h.ComicDao.IncrChapterCountByComicIDWithSession(ctx, session, task.ComicId)
		if err != nil {
			return err
		}
		// 2. 添加章节图片信息
		comicPageModels := convert.ComicPagesFromStrings(task.ChapterID, task.PageUrls)
		err = h.ComicPageDao.InsertBatchWithSession(ctx, session, comicPageModels)
		if err != nil {
			return err
		}
		// 3. 更新章节的关联关系
		return h.ComicChapterDao.UpdateRelationStatusWithSession(ctx, session, task.ChapterID, logic.RelationStatusNormal)
	})
}

func (h *ComicChapterRelationHandler) handleModify(ctx context.Context, task *tasks.ComicChapterRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 删除原有章节图片信息
		err := h.ComicPageDao.DeleteByChapterIDWithSession(ctx, session, task.ChapterID)
		if err != nil {
			return err
		}

		// 2. 添加新的章节图片信息
		comicPageModels := convert.ComicPagesFromStrings(task.ChapterID, task.PageUrls)
		err = h.ComicPageDao.InsertBatchWithSession(ctx, session, comicPageModels)
		if err != nil {
			return err
		}

		// 3. 更新章节的关联关系
		return h.ComicChapterDao.UpdateRelationStatusWithSession(ctx, session, task.ChapterID, logic.RelationStatusNormal)
	})
}

func (h *ComicChapterRelationHandler) handleDelete(ctx context.Context, task *tasks.ComicChapterRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		err := h.ComicDao.DecrChapterCountByComicIDWithSession(ctx, session, task.ComicId)
		if err != nil {
			return err
		}
		return h.ComicPageDao.DeleteByChapterIDWithSession(ctx, session, task.ChapterID)
	})
}

func (h *ComicChapterRelationHandler) handleDeleteAll(ctx context.Context, task *tasks.ComicChapterRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 获取该漫画的所有章节
		chapters, err := h.ComicChapterDao.FindAllByComicID(ctx, task.ComicId)
		if err != nil {
			return err
		}

		// 2. 更新每一个章节的删除状态
		ids := make([]uint64, len(chapters))
		for i, chapter := range chapters {
			ids[i] = chapter.Id
		}
		err = h.ComicChapterDao.SoftDeleteByIDsWithSession(ctx, session, ids, int64(task.UID))
		if err != nil {
			return err
		}

		// 3. 删除每一个章节的图片信息
		return h.ComicPageDao.DeleteAllByChapterIDsWithSession(ctx, session, ids)
	})
}
