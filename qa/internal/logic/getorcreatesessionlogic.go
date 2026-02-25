package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"
	"github.com/luyb177/XiaoAnBackend/qa/internal/model"
	"github.com/luyb177/XiaoAnBackend/qa/internal/svc"
	"github.com/luyb177/XiaoAnBackend/qa/pb/qa/v1"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrCreateSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ChatSessionDao model.ChatSessionModel
}

func NewGetOrCreateSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrCreateSessionLogic {
	return &GetOrCreateSessionLogic{
		ctx:            ctx,
		svcCtx:         svcCtx,
		Logger:         logx.WithContext(ctx),
		ChatSessionDao: model.NewChatSessionModel(svcCtx.Mysql),
	}
}

// GetOrCreateSession 获取新会话
func (l *GetOrCreateSessionLogic) GetOrCreateSession(in *v1.GetOrCreateSessionRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || user.Role == "" || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或状态异常"), nil
	}

	// 事务
	var chatSession *model.ChatSession
	err := l.svcCtx.Mysql.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 查询是否有空的对话
		var err error
		chatSession, err = l.ChatSessionDao.FindEmptySessionByUserIDWithSession(ctx, session, user.UID)
		if err == nil {
			return nil
		} else if !errors.Is(err, model.ErrNotFound) {
			return err
		}

		// 2. 没有空的对话，创建一个新的对话
		now := time.Now()
		chatSession = &model.ChatSession{
			Title:          "新对话",
			UserId:         user.UID,
			MessageCount:   0,
			EmptySlot:      sql.NullInt64{Int64: 1, Valid: true}, // 预占一个空位，防止并发创建多个空对话
			HasMessage:     HasMessageNo,
			SessionStatus:  SessionStatusEmpty,
			IsPinned:       PinnedNo,
			PinnedAt:       sql.NullTime{},
			LastMessageAt:  now,
			RelationStatus: RelationStatusNormal,
			CreatedAt:      now,
			UpdatedAt:      now,
			DeletedAt:      0,
		}

		result, err := l.ChatSessionDao.InsertWithSession(ctx, session, chatSession)
		if err != nil {
			if errors.Is(err, model.ErrDuplicateEntry) {
				// 可能是并发创建了多个空对话，直接查询
				chatSession, err = l.ChatSessionDao.FindEmptySessionByUserIDWithSession(ctx, session, user.UID)
				return err
			}
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		chatSession.Id = uint64(id)
		return nil
	})

	if err != nil {
		l.Errorf("GetOrCreateSession err: %v", err)
		return bad("获取或创建对话失败"), nil
	}

	// 尽力加到 redis 缓存
	err = l.svcCtx.ChatSessionRepo.SetSessionOwner(l.ctx, chatSession.Id, user.UID)
	if err != nil {
		l.Errorf("GetOrCreateSession SetSessionOwner err: %v", err)
	}

	res := &v1.GetOrCreateSessionResponse{Session: &v1.ChatSession{
		Id:             chatSession.Id,
		Title:          chatSession.Title,
		UserId:         chatSession.UserId,
		MessageCount:   chatSession.MessageCount,
		HasMessage:     chatSession.HasMessage,
		SessionStatus:  chatSession.SessionStatus,
		IsPinned:       chatSession.IsPinned,
		PinnedAt:       chatSession.PinnedAt.Time.Unix(),
		LastMessageAt:  chatSession.LastMessageAt.Unix(),
		RelationStatus: chatSession.RelationStatus,
		CreatedAt:      chatSession.CreatedAt.Unix(),
		UpdatedAt:      chatSession.UpdatedAt.Unix(),
	}}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("GetOrCreateSession err: %v", err)
		return internal("封装返回结果失败"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "获取或创建对话成功",
		Data:    resAny,
	}, nil
}
