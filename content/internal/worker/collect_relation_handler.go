package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/luyb177/XiaoAnBackend/content/internal/logic"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
	"github.com/luyb177/XiaoAnBackend/infra/queue"
	"github.com/luyb177/XiaoAnBackend/infra/queue/redisqueue"
)

type CollectRelationHandler struct {
	logx.Logger
	svcCtx            *svc.ServiceContext
	ContentCollectDao model.ContentCollectModel

	ArticleDao model.ArticleModel
	PodcastDao model.PodcastModel
	VideoDao   model.VideoModel
	ComicDao   model.ComicModel
}

func NewCollectRelationHandler(ctx context.Context, svcCtx *svc.ServiceContext) *CollectRelationHandler {
	return &CollectRelationHandler{
		svcCtx:            svcCtx,
		Logger:            logx.WithContext(ctx),
		ContentCollectDao: model.NewContentCollectModel(svcCtx.Mysql),
		ArticleDao:        model.NewArticleModel(svcCtx.Mysql),
		PodcastDao:        model.NewPodcastModel(svcCtx.Mysql),
		VideoDao:          model.NewVideoModel(svcCtx.Mysql),
		ComicDao:          model.NewComicModel(svcCtx.Mysql),
	}
}

func (h *CollectRelationHandler) Handle(ctx context.Context, task queue.Task) error {
	payload, err := task.Payload()
	if err != nil {
		return err
	}

	var rawTask redisqueue.RawTask
	err = json.Unmarshal(payload, &rawTask)
	if err != nil {
		return err
	}

	collectTask := tasks.CollectRelationTask{}
	err = json.Unmarshal(rawTask.Data, &collectTask)
	if err != nil {
		return err
	}

	h.Infof("CollectRelationHandler received task: %+v", collectTask)

	switch collectTask.Type {
	case tasks.CollectRelationAdd:
		return h.handleAdd(ctx, &collectTask)
	case tasks.CollectRelationDelete:
		return h.handleDelete(ctx, &collectTask)
	default:
		h.Errorf("CollectRelationHandler unknown task type: %s", collectTask.Type)
		return nil
	}
}

func (h *CollectRelationHandler) handleAdd(ctx context.Context, task *tasks.CollectRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		contentCollect := &model.ContentCollect{
			Type:      task.ContentType,
			TargetId:  task.ContentID,
			UserId:    task.UID,
			IsCounted: 0,
			DeletedAt: 0,
		}
		// 确保存在性
		_, err := h.ContentCollectDao.UpsertWithSession(ctx, session, contentCollect)
		if err != nil {
			return err
		}
		// 获取 ID
		contentCollect, err = h.ContentCollectDao.FindOneByUserIDTypeTargetIDWithSession(ctx, session, task.UID, task.ContentType, task.ContentID)
		if err != nil {
			return err
		}

		// 增加对应的收藏计数 标记
		result, err := h.ContentCollectDao.MarkCollectAsCountedWithSession(ctx, session, contentCollect.Id)
		if err != nil {
			return err
		}
		affect, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affect == 0 {
			// 已经计数过，直接返回
			return nil
		}

		// 增加对应内容的收藏计数
		result, err = h.incrContentCollectCount(ctx, session, task.ContentType, task.ContentID)
		if err != nil {
			return err
		}
		affect, err = result.RowsAffected()
		if err != nil {
			return err
		}
		if affect == 0 {
			// 内容不存在，回滚收藏计数标记,删除 collect 记录
			_, err = h.ContentCollectDao.UnmarkCollectAsCountedWithSession(ctx, session, contentCollect.Id)
			if err != nil {
				return err
			}
			_, err = h.ContentCollectDao.SoftDeleteWithSession(ctx, session, contentCollect.Id, uint64(time.Now().Unix()))
			return err
		}
		return nil
	})
}

func (h *CollectRelationHandler) handleDelete(ctx context.Context, task *tasks.CollectRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 先 find
		// todo : 可以不用 find ，避免 toctou 问题
		contentCollect, err := h.ContentCollectDao.FindOneByUserIDTypeTargetIDWithSession(ctx, session, task.UID, task.ContentType, task.ContentID)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil
			}
			return err
		}

		// 标记
		result, err := h.ContentCollectDao.UnmarkCollectAsCountedWithSession(ctx, session, contentCollect.Id)
		if err != nil {
			return err
		}
		affect, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affect > 0 {
			// 计数过
			// 内容 collect - 1
			result, err = h.decrContentCollectCount(ctx, session, task.ContentType, task.ContentID)
			if err != nil {
				return err
			}
			affect, err = result.RowsAffected()
			if err != nil {
				return err
			}
			if affect == 0 {
				// 内容不存在
				h.Errorf("CollectRelationHandler handleDelete: content not found, contentType: %s, contentID: %d", task.ContentType, task.ContentID)
			}
		}

		// 删除
		_, err = h.ContentCollectDao.SoftDeleteWithSession(ctx, session, contentCollect.Id, uint64(time.Now().Unix()))
		return err
	})
}

func (h *CollectRelationHandler) incrContentCollectCount(ctx context.Context, session sqlx.Session, contentType string, contentID uint64) (sql.Result, error) {
	var result sql.Result
	var err error

	switch contentType {
	case logic.ContentTypeArticle:
		result, err = h.ArticleDao.IncrCollectCountWithSession(ctx, session, contentID)
	case logic.ContentTypePodcast:
		result, err = h.PodcastDao.IncrCollectCountWithSession(ctx, session, contentID)
	case logic.ContentTypeVideo:
		result, err = h.VideoDao.IncrCollectCountWithSession(ctx, session, contentID)
	case logic.ContentTypeComic:
		result, err = h.ComicDao.IncrCollectCountWithSession(ctx, session, contentID)
	default:
		return nil, errors.New("unknown content type")
	}
	return result, err
}

func (h *CollectRelationHandler) decrContentCollectCount(ctx context.Context, session sqlx.Session, contentType string, contentID uint64) (sql.Result, error) {
	var result sql.Result
	var err error
	switch contentType {
	case logic.ContentTypeArticle:
		result, err = h.ArticleDao.DecrCollectCountWithSession(ctx, session, contentID)
	case logic.ContentTypePodcast:
		result, err = h.PodcastDao.DecrCollectCountWithSession(ctx, session, contentID)
	case logic.ContentTypeVideo:
		result, err = h.VideoDao.DecrCollectCountWithSession(ctx, session, contentID)
	case logic.ContentTypeComic:
		result, err = h.ComicDao.DecrCollectCountWithSession(ctx, session, contentID)
	default:
		return nil, errors.New("unknown content type")
	}
	return result, err
}
