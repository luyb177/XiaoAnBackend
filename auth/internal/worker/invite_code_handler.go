package worker

import (
	"context"
	"encoding/json"

	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	"github.com/luyb177/XiaoAnBackend/auth/internal/repo/redisqueue"
	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pkg/taskqueue"
	"github.com/luyb177/XiaoAnBackend/auth/pkg/taskqueue/tasks"

	"github.com/zeromicro/go-zero/core/logx"
)

type InviteRelationHandler struct {
	logx.Logger
	svcCtx        *svc.ServiceContext
	InviteCodeDao model.InviteCodeModel
}

func NewInviteRelationHandler(svcCtx *svc.ServiceContext, ctx context.Context) *InviteRelationHandler {
	return &InviteRelationHandler{
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		InviteCodeDao: model.NewInviteCodeModel(svcCtx.Mysql),
	}
}

func (h *InviteRelationHandler) Handle(ctx context.Context, task taskqueue.Task) error {
	payload, err := task.Payload()
	if err != nil {
		return err
	}
	var rawTask redisqueue.RawTask
	err = json.Unmarshal(payload, &rawTask)
	if err != nil {
		return err
	}

	inviteCodeTask := tasks.InviteCodeTask{}
	err = json.Unmarshal(rawTask.Data, &inviteCodeTask)
	if err != nil {
		return err
	}

	h.Infof("processing invite code task: %+v", inviteCodeTask)
	switch inviteCodeTask.Type {
	case tasks.InviteCodeRelationNotActive:
		return h.handleNotActive(ctx, &inviteCodeTask)
	default:
		h.Errorf("unknown invite code task: %s", inviteCodeTask.Type)
		return nil
	}

}

func (h *InviteRelationHandler) handleNotActive(ctx context.Context, task *tasks.InviteCodeTask) error {
	_, err := h.InviteCodeDao.NotActiveByCode(ctx, task.Code)
	return err
}
