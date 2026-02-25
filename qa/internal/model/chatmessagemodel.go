package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ChatMessageModel = (*customChatMessageModel)(nil)

type (
	// ChatMessageModel is an interface to be customized, add more methods here,
	// and implement the added methods in customChatMessageModel.
	ChatMessageModel interface {
		chatMessageModel
		withSession(session sqlx.Session) ChatMessageModel
		Insert(ctx context.Context, data *ChatMessage) (sql.Result, error)
		InsertWithSession(ctx context.Context, session sqlx.Session, data *ChatMessage) (sql.Result, error)
		FindMessagesByChatSessionID(ctx context.Context, sessionID uint64, limit int64) ([]*ChatMessage, error)
		FindOneBySessionIdMessageId(ctx context.Context, sessionID uint64, messageID string) (*ChatMessage, error)
		FindOneBySessionIdMessageIdWithSession(ctx context.Context, session sqlx.Session, sessionID uint64, messageID string) (*ChatMessage, error)
		UpdateContent(ctx context.Context, data *ChatMessage) (sql.Result, error)
		UpdateContentFinishedSuccess(ctx context.Context, data *ChatMessage) (sql.Result, error)
		UpdateContentFinishedFailed(ctx context.Context, data *ChatMessage) (sql.Result, error)
		FindManyByUserIDSessionID(ctx context.Context, userID, sessionID uint64, limit int64) ([]*ChatMessage, error)
		FindManyByUserIDSessionIDWithCursor(ctx context.Context, userID, sessionID, cursorID uint64, limit int64) ([]*ChatMessage, error)
	}

	customChatMessageModel struct {
		*defaultChatMessageModel
	}
)

// NewChatMessageModel returns a model for the database table.
func NewChatMessageModel(conn sqlx.SqlConn) ChatMessageModel {
	return &customChatMessageModel{
		defaultChatMessageModel: newChatMessageModel(conn),
	}
}

func (m *customChatMessageModel) withSession(session sqlx.Session) ChatMessageModel {
	return NewChatMessageModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customChatMessageModel) Insert(ctx context.Context, data *ChatMessage) (sql.Result, error) {
	//nolint:staticcheck // QF1008: go-zero embedding style retained intentionally
	result, err := m.defaultChatMessageModel.Insert(ctx, data)
	return result, mapDBError(err)
}

func (m *customChatMessageModel) InsertWithSession(ctx context.Context, session sqlx.Session, data *ChatMessage) (sql.Result, error) {
	return m.withSession(session).Insert(ctx, data)
}

func (m *customChatMessageModel) FindMessagesByChatSessionID(ctx context.Context, sessionID uint64, limit int64) ([]*ChatMessage, error) {
	query := fmt.Sprintf(`
		select %s from %s
		where session_id = ?
			and deleted_at = 0
			order by id asc 
			limit ?`,
		chatMessageRows,
		m.table,
	)

	var messages []*ChatMessage
	err := m.conn.QueryRowsCtx(ctx, &messages, query, sessionID, limit)
	return messages, mapDBError(err)
}

func (m *customChatMessageModel) UpdateContent(ctx context.Context, data *ChatMessage) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s set content = ?
		where id = ?
		    and status = 0
		    and deleted_at = 0`,
		m.table,
	)
	result, err := m.conn.ExecCtx(ctx, query, data.Content, data.Id)
	return result, mapDBError(err)
}

func (m *customChatMessageModel) FindOneBySessionIdMessageId(ctx context.Context, sessionID uint64, messageID string) (*ChatMessage, error) {
	// nolint:staticcheck // QF1008: go-zero embedding style retained intentionally
	resp, err := m.defaultChatMessageModel.FindOneBySessionIdMessageId(ctx, sessionID, messageID)
	return resp, mapDBError(err)
}

func (m *customChatMessageModel) FindOneBySessionIdMessageIdWithSession(ctx context.Context, session sqlx.Session, sessionID uint64, messageID string) (*ChatMessage, error) {
	return m.withSession(session).FindOneBySessionIdMessageId(ctx, sessionID, messageID)
}

func (m *customChatMessageModel) UpdateContentFinishedSuccess(ctx context.Context, data *ChatMessage) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s set content = ?,
		              status = 1,
		              prompt_tokens = ?,
		              completion_tokens = ?,
		              total_tokens = ?,
		              finish_reason = ?
	  	where id = ?
	  		and deleted_at = 0
	  		and status = 0`,
		m.table,
	)
	result, err := m.conn.ExecCtx(ctx, query, data.Content, data.PromptTokens, data.CompletionTokens, data.TotalTokens, data.FinishReason, data.Id)
	return result, mapDBError(err)
}

func (m *customChatMessageModel) UpdateContentFinishedFailed(ctx context.Context, data *ChatMessage) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s set content = ?,
		              status = 2,
		              prompt_tokens = ?,
		              completion_tokens = ?,
		              total_tokens = ?,
		              finish_reason = ?
	  	where id = ?
	  		and deleted_at = 0
	  		and status = 0`,
		m.table,
	)
	result, err := m.conn.ExecCtx(ctx, query, data.Content, data.PromptTokens, data.CompletionTokens, data.TotalTokens, data.FinishReason, data.Id)
	return result, mapDBError(err)
}

func (m *customChatMessageModel) FindManyByUserIDSessionID(ctx context.Context, userID, sessionID uint64, limit int64) ([]*ChatMessage, error) {
	query := fmt.Sprintf(`
		select %s from %s
		where user_id = ?
			and session_id = ?
			and status = 1
			and deleted_at = 0
			order by id desc 
			limit ?`,
		chatMessageRows,
		m.table,
	)

	var chatMessages []*ChatMessage
	err := m.conn.QueryRowsCtx(ctx, &chatMessages, query, userID, sessionID, limit)
	return chatMessages, mapDBError(err)
}

func (m *customChatMessageModel) FindManyByUserIDSessionIDWithCursor(ctx context.Context, userID, sessionID, cursorID uint64, limit int64) ([]*ChatMessage, error) {
	query := fmt.Sprintf(`
		select %s from %s
		where user_id = ?
			and session_id = ?
			and id < ?
			and status = 1
			and deleted_at = 0
			order by id desc 
			limit ?`,
		chatMessageRows,
		m.table,
	)

	var chatMessages []*ChatMessage
	err := m.conn.QueryRowsCtx(ctx, &chatMessages, query, userID, sessionID, cursorID, limit)
	return chatMessages, mapDBError(err)
}
