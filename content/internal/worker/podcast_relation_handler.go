package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/luyb177/XiaoAnBackend/content/internal/logic"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/repo/redisqueue"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pkg/podcast/convert"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
)

type PodcastRelationHandler struct {
	logx.Logger
	svcCtx              *svc.ServiceContext
	PodcastDao          model.PodcastModel
	PodcastTagDao       model.PodcastTagModel
	PodcastHighlightDao model.PodcastHighlightModel

	CommentDao     model.CommentModel
	ContentLikeDao model.ContentLikeModel
}

func NewPodcastRelationHandler(svcCtx *svc.ServiceContext, ctx context.Context) *PodcastRelationHandler {
	return &PodcastRelationHandler{
		svcCtx:              svcCtx,
		Logger:              logx.WithContext(ctx),
		PodcastDao:          model.NewPodcastModel(svcCtx.Mysql),
		PodcastTagDao:       model.NewPodcastTagModel(svcCtx.Mysql),
		PodcastHighlightDao: model.NewPodcastHighlightModel(svcCtx.Mysql),
		CommentDao:          model.NewCommentModel(svcCtx.Mysql),
		ContentLikeDao:      model.NewContentLikeModel(svcCtx.Mysql),
	}
}

func (h *PodcastRelationHandler) Handle(ctx context.Context, task taskqueue.Task) error {
	payload, err := task.Payload()
	if err != nil {
		return err
	}

	var rawTask redisqueue.RawTask
	err = json.Unmarshal(payload, &rawTask)
	if err != nil {
		return err
	}

	podcastTask := tasks.PodcastRelationTask{}
	err = json.Unmarshal(rawTask.Data, &podcastTask)
	if err != nil {
		return err
	}

	h.Infof("processing podcast relation task: %+v", podcastTask)

	switch podcastTask.Type {
	case tasks.PodcastRelationAdd:
		return h.handleAdd(ctx, &podcastTask)
	case tasks.PodcastRelationModify:
		return h.handleModify(ctx, &podcastTask)
	case tasks.PodcastRelationDelete:
		return h.handleDelete(ctx, &podcastTask)
	case tasks.PodcastRelationGet:
		return h.handleGet(ctx, &podcastTask)
	default:
		h.Errorf("unknown task type: %s", podcastTask.Type)
		return nil
	}
}

func (h *PodcastRelationHandler) handleAdd(ctx context.Context, task *tasks.PodcastRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 添加标签
		tagModels := convert.PodcastTagsFromStrings(task.PodcastID, task.Tags)
		err := h.PodcastTagDao.InsertBatchWithSession(ctx, session, tagModels)
		if err != nil {
			return err
		}

		// 2. 添加重点时间
		highlightModels := convert.PodcastHighlightsFromPB(task.PodcastID, task.Highlights)
		err = h.PodcastHighlightDao.InsertBatchWithSession(ctx, session, highlightModels)
		if err != nil {
			return err
		}

		// 3. 更新关联状态
		return h.PodcastDao.UpdateRelationStatusWithSession(ctx, session, task.PodcastID, logic.RelationStatusNormal)
	})
}
func (h *PodcastRelationHandler) handleModify(ctx context.Context, task *tasks.PodcastRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 删除旧标签
		err := h.PodcastTagDao.DeleteBatchByPodcastIdWithSession(ctx, session, task.PodcastID)
		if err != nil {
			return err
		}

		// 2. 添加新标签
		tagModels := convert.PodcastTagsFromStrings(task.PodcastID, task.Tags)
		err = h.PodcastTagDao.InsertBatchWithSession(ctx, session, tagModels)
		if err != nil {
			return err
		}

		// 3. 删除旧重点时间
		err = h.PodcastHighlightDao.DeleteBatchByPodcastIdWithSession(ctx, session, task.PodcastID)
		if err != nil {
			return err
		}

		// 4. 添加新重点时间
		highlightModels := convert.PodcastHighlightsFromPB(task.PodcastID, task.Highlights)
		err = h.PodcastHighlightDao.InsertBatchWithSession(ctx, session, highlightModels)
		if err != nil {
			return err
		}

		// 5. 更新关联状态
		return h.PodcastDao.UpdateRelationStatusWithSession(ctx, session, task.PodcastID, logic.RelationStatusNormal)
	})
}
func (h *PodcastRelationHandler) handleDelete(ctx context.Context, task *tasks.PodcastRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		deletedAt := uint64(time.Now().Unix())
		// 1. 删除标签
		err := h.PodcastTagDao.SoftDeleteByPodcastIdWithSession(ctx, session, task.PodcastID, deletedAt)
		if err != nil {
			return err
		}

		// 2. 删除重点时间
		err = h.PodcastHighlightDao.SoftDeleteByPodcastIdWithSession(ctx, session, task.PodcastID, deletedAt)
		if err != nil {
			return err
		}

		// 3. 删除评论
		_, err = h.CommentDao.SoftDeleteByTypeAndTargetIdWithSession(ctx, session, logic.ContentTypePodcast, task.PodcastID, deletedAt)
		if err != nil {
			return err
		}

		// 删除点赞
		_, err = h.ContentLikeDao.SoftDeleteByTypeTargetIdWithSession(ctx, session, logic.ContentTypePodcast, task.PodcastID, deletedAt)
		return err
	})
}

func (h *PodcastRelationHandler) handleGet(ctx context.Context, task *tasks.PodcastRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 增加播客浏览量
		_, err := h.PodcastDao.IncrViewCountWithSession(ctx, session, task.PodcastID)
		return err
	})
}
