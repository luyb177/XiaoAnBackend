package worker

import (
	"context"
	"encoding/json"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/luyb177/XiaoAnBackend/auth/internal/repo/redisqueue"
	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pkg/email"
	"github.com/luyb177/XiaoAnBackend/auth/pkg/taskqueue"
	"github.com/luyb177/XiaoAnBackend/auth/pkg/taskqueue/tasks"
)

type EmailRelationHandler struct {
	logx.Logger
	svcCtx      *svc.ServiceContext
	EmailConfig *email.Config
}

func NewEmailRelationHandler(svcCtx *svc.ServiceContext, ctx context.Context) *EmailRelationHandler {
	emailConfig := &email.Config{
		From:     svcCtx.Config.Email.From,
		Password: svcCtx.Config.Email.Password,
		SMTPHost: svcCtx.Config.Email.SMTPHost,
		SMTPPort: svcCtx.Config.Email.SMTPPort,
	}

	return &EmailRelationHandler{
		svcCtx:      svcCtx,
		Logger:      logx.WithContext(ctx),
		EmailConfig: emailConfig,
	}
}

func (h *EmailRelationHandler) Handle(ctx context.Context, task taskqueue.Task) error {
	payload, err := task.Payload()
	if err != nil {
		return err
	}
	var rawTask redisqueue.RawTask
	err = json.Unmarshal(payload, &rawTask)
	if err != nil {
		return err
	}

	emailTask := tasks.EmailRelationTask{}
	err = json.Unmarshal(rawTask.Data, &emailTask)
	if err != nil {
		return err
	}

	h.Infof("processing email relation task: %+v", emailTask)
	switch emailTask.Type {
	case tasks.EmailRelationSend:
		return h.handleSend(ctx, &emailTask)
	case tasks.EmailCodeRelationDelete:
		return h.handleDelete(ctx, &emailTask)
	default:
		h.Errorf("unknown email relation task type: %s", emailTask.Type)
		return nil
	}
}

func (h *EmailRelationHandler) handleSend(ctx context.Context, task *tasks.EmailRelationTask) error {
	return email.SendEmailCode(h.EmailConfig, task.To, task.Code)
}

func (h *EmailRelationHandler) handleDelete(ctx context.Context, task *tasks.EmailRelationTask) error {
	return h.svcCtx.RedisRepo.EmailRepo.DelEmailCode(task.To)
}
