package logic

import (
	"context"
	"time"

	"github.com/luyb177/XiaoAnBackend/auth/internal/svc"
	"github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
	authcode "github.com/luyb177/XiaoAnBackend/auth/pkg/code"
	"github.com/luyb177/XiaoAnBackend/auth/pkg/email"
	"github.com/luyb177/XiaoAnBackend/auth/pkg/taskqueue/tasks"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendEmailCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendEmailCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendEmailCodeLogic {

	return &SendEmailCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SendEmailCode 邮箱验证码
func (l *SendEmailCodeLogic) SendEmailCode(in *v1.SendEmailRequest) (*v1.Response, error) {
	if in.Email == "" {
		return bad("邮箱不能为空"), nil
	}

	// 基本验证
	// 	1. 长度
	if len(in.Email) > 254 {
		return bad("邮箱长度不能超过254个字符"), nil
	}

	// 	2. trim spaces and to lower
	in.Email = email.CanonicalEmail(in.Email)

	// 	3. 基本格式验证
	if !email.IsValidEmail(in.Email) {
		return bad("邮箱格式不正确"), nil
	}

	// 生成验证码
	code := authcode.EmailCode()

	// 先存储 后发送
	err := l.svcCtx.RedisRepo.EmailRepo.SetEmailCode(in.Email, code, time.Minute*5)
	if err != nil {
		l.Errorf("设置邮件验证码失败: %v", err)
		return bad("设置邮件验证码失败"), nil
	}

	// 加入任务队列 异步发送邮件
	emailRelationTask := &tasks.EmailRelationTask{
		Type: tasks.EmailRelationSend,
		To:   in.Email,
		Code: code,
	}

	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, emailRelationTask)
	if err != nil {
		// 加入任务队列失败 不影响用户使用 只是无法发送邮件
		l.Errorf("加入邮件任务队列失败: %v", err)
	}

	return &v1.Response{
		Code:    200,
		Message: "邮件发送中，请注意查收",
	}, nil
}
