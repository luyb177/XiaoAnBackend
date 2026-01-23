package worker

import (
	"context"
	"encoding/json"

	"github.com/luyb177/XiaoAnBackend/content/internal/logic"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/repo/redisqueue"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
	"github.com/luyb177/XiaoAnBackend/content/pkg/video/convert"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type VideoRelationHandler struct {
	logx.Logger
	svcCtx      *svc.ServiceContext
	VideoDao    model.VideoModel
	VideoTagDao model.VideoTagModel
}

func NewVideoRelationHandler(svcCtx *svc.ServiceContext, ctx context.Context) *VideoRelationHandler {
	return &VideoRelationHandler{
		svcCtx:      svcCtx,
		Logger:      logx.WithContext(ctx),
		VideoDao:    model.NewVideoModel(svcCtx.Mysql),
		VideoTagDao: model.NewVideoTagModel(svcCtx.Mysql),
	}
}

func (h *VideoRelationHandler) Handle(ctx context.Context, task taskqueue.Task) error {
	payload, err := task.Payload()
	if err != nil {
		return err
	}

	var rawTask redisqueue.RawTask
	err = json.Unmarshal(payload, &rawTask)
	if err != nil {
		return err
	}

	var videoTask tasks.VideoRelationTask
	err = json.Unmarshal(rawTask.Data, &videoTask)
	if err != nil {
		return err
	}

	h.Infof("processing video relation task: %+v", videoTask)

	switch videoTask.Type {
	case tasks.VideoRelationAdd:
		return h.handleAdd(ctx, &videoTask)
	case tasks.VideoRelationModify:
		return h.handleModify(ctx, &videoTask)
	case tasks.VideoRelationDelete:
		return h.handleDelete(ctx, &videoTask)
	default:
		h.Errorf("unknown task type: %s", videoTask.Type)
		return nil
	}
}

func (h *VideoRelationHandler) handleAdd(ctx context.Context, task *tasks.VideoRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 添加标签
		tagModels := convert.VideoTagsFromStrings(task.VideoID, task.Tags)
		err := h.VideoTagDao.InsertBatchWithSession(ctx, session, tagModels)
		if err != nil {
			return err
		}

		// 2. 更新关联状态
		return h.VideoDao.UpdateRelationStatusWithSession(ctx, session, task.VideoID, logic.RelationStatusNormal)
	})
}

func (h *VideoRelationHandler) handleModify(ctx context.Context, task *tasks.VideoRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 删除旧标签
		err := h.VideoTagDao.DeleteBatchByVideoIdWithSession(ctx, session, task.VideoID)
		if err != nil {
			return err
		}

		// 2. 插入新标签
		tagsModel := convert.VideoTagsFromStrings(task.VideoID, task.Tags)
		err = h.VideoTagDao.InsertBatchWithSession(ctx, session, tagsModel)
		if err != nil {
			return err
		}

		// 3. 更新关联状态
		return h.VideoDao.UpdateRelationStatusWithSession(ctx, session, task.VideoID, logic.RelationStatusNormal)
	})
}

func (h *VideoRelationHandler) handleDelete(ctx context.Context, task *tasks.VideoRelationTask) error {
	return h.VideoTagDao.DeleteBatchByVideoId(ctx, task.VideoID)
}
