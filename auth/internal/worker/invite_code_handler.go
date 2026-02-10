package worker

import (
	"context"
	"encoding/json"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	"github.com/luyb177/XiaoAnBackend/auth/internal/repo/redisqueue"
	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pkg/taskqueue"
	"github.com/luyb177/XiaoAnBackend/auth/pkg/taskqueue/tasks"
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
	default:
		h.Errorf("unknown invite code task: %s", inviteCodeTask.Type)
		return nil
	}

}
