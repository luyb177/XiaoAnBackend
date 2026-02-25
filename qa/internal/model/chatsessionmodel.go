package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ChatSessionModel = (*customChatSessionModel)(nil)

type (
	// ChatSessionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customChatSessionModel.
	ChatSessionModel interface {
		chatSessionModel
		withSession(session sqlx.Session) ChatSessionModel
		Insert(ctx context.Context, data *ChatSession) (sql.Result, error)
		InsertWithSession(ctx context.Context, session sqlx.Session, data *ChatSession) (sql.Result, error)
		FindEmptySessionByUserID(ctx context.Context, userID uint64) (*ChatSession, error)
		FindEmptySessionByUserIDWithSession(ctx context.Context, session sqlx.Session, userID uint64) (*ChatSession, error)
		FindOneByIDUserID(ctx context.Context, id uint64, userID uint64) (*ChatSession, error)
		FindEmptyOneByIDUerID(ctx context.Context, id, userID uint64) (*ChatSession, error)
		UpdateEmptySlot(ctx context.Context, id uint64, lastMessageAt time.Time) (sql.Result, error)
		IncrementMessageCount(ctx context.Context, id uint64, lastMessageAt time.Time) (sql.Result, error)
		IncrementMessageCountWithSession(ctx context.Context, session sqlx.Session, id uint64, lastMessageAt time.Time) (sql.Result, error)
		UpdateTitle(ctx context.Context, id uint64, title string) (sql.Result, error)
		FindManyByUserID(ctx context.Context, userID uint64, limit int64) ([]*ChatSession, error)
		FindManyByUserIDWithCursor(ctx context.Context, userID uint64, cursorTime time.Time, cursorID uint64, limit int64) ([]*ChatSession, error)
	}

	customChatSessionModel struct {
		*defaultChatSessionModel
	}
)

// NewChatSessionModel returns a model for the database table.
func NewChatSessionModel(conn sqlx.SqlConn) ChatSessionModel {
	return &customChatSessionModel{
		defaultChatSessionModel: newChatSessionModel(conn),
	}
}

func (m *customChatSessionModel) withSession(session sqlx.Session) ChatSessionModel {
	return NewChatSessionModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customChatSessionModel) Insert(ctx context.Context, data *ChatSession) (sql.Result, error) {
	//nolint:staticcheck // QF1008: go-zero embedding style retained intentionally
	result, err := m.defaultChatSessionModel.Insert(ctx, data)
	return result, mapDBError(err)
}

func (m *customChatSessionModel) InsertWithSession(ctx context.Context, session sqlx.Session, data *ChatSession) (sql.Result, error) {
	return m.withSession(session).Insert(ctx, data)
}

func (m *customChatSessionModel) FindEmptySessionByUserID(ctx context.Context, userID uint64) (*ChatSession, error) {
	query := fmt.Sprintf(`
		select %s from %s
		where user_id = ?
			and empty_slot = 1
			and deleted_at = 0
		limit 1
		for update`,
		chatSessionRows,
		m.table,
	)

	var resp ChatSession
	err := m.conn.QueryRowCtx(ctx, &resp, query, userID)
	return &resp, mapDBError(err)
}

func (m *customChatSessionModel) FindEmptySessionByUserIDWithSession(ctx context.Context, session sqlx.Session, userID uint64) (*ChatSession, error) {
	return m.withSession(session).FindEmptySessionByUserID(ctx, userID)
}

func (m *customChatSessionModel) FindOneByIDUserID(ctx context.Context, id, userID uint64) (*ChatSession, error) {
	query := fmt.Sprintf(`
		select %s from %s
		where id = ? 
		    and user_id = ? 
		    and deleted_at = 0
		limit 1`,
		chatSessionRows,
		m.table,
	)

	var resp ChatSession
	err := m.conn.QueryRowCtx(ctx, &resp, query, id, userID)
	return &resp, mapDBError(err)
}

func (m *customChatSessionModel) FindEmptyOneByIDUerID(ctx context.Context, id, userID uint64) (*ChatSession, error) {
	query := fmt.Sprintf(`
		select %s from %s
		where id = ?
			and user_id = ?
			and empty_slot = 1
			and deleted_at = 0
		limit 1`,
		chatSessionRows,
		m.table,
	)

	var resp ChatSession
	err := m.conn.QueryRowCtx(ctx, &resp, query, id, userID)
	return &resp, mapDBError(err)
}

func (m *customChatSessionModel) FindManyByUserID(ctx context.Context, userID uint64, limit int64) ([]*ChatSession, error) {
	query := fmt.Sprintf(`
		select %s from %s
		where user_id = ?
			and empty_slot is null
			and is_pinned = 0
			and deleted_at = 0
		order by last_message_at desc,id desc
		limit ?`,
		chatSessionRows,
		m.table)

	var resp []*ChatSession
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userID, limit)
	return resp, mapDBError(err)
}

func (m *customChatSessionModel) FindManyByUserIDWithCursor(ctx context.Context, userID uint64, cursorTime time.Time, cursorID uint64, limit int64) ([]*ChatSession, error) {
	query := fmt.Sprintf(`
		select %s from %s
		where user_id = ?
			and empty_slot is null
			and is_pinned = 0
			and deleted_at = 0
			and (last_message_at < ? or (last_message_at = ? and id < ?))
		order by last_message_at desc, id desc
		limit ?`,
		chatSessionRows,
		m.table)

	var resp []*ChatSession
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userID, cursorTime, cursorTime, cursorID, limit)
	return resp, mapDBError(err)
}

func (m *customChatSessionModel) UpdateEmptySlot(ctx context.Context, id uint64, lastMessageAt time.Time) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set empty_slot = null, 
		    session_status = 1,
			last_message_at = ?
		where id = ?
			and empty_slot = 1
			and deleted_at = 0`,
		m.table,
	)

	result, err := m.conn.ExecCtx(ctx, query, lastMessageAt, id)
	return result, mapDBError(err)
}

func (m *customChatSessionModel) IncrementMessageCount(ctx context.Context, id uint64, lastMessageAt time.Time) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set message_count = message_count + 1,
		    has_message = 1,
		    last_message_at = ?
		where id = ?
			and deleted_at = 0`,
		m.table,
	)

	result, err := m.conn.ExecCtx(ctx, query, lastMessageAt, id)
	return result, mapDBError(err)
}

func (m *customChatSessionModel) IncrementMessageCountWithSession(ctx context.Context, session sqlx.Session, id uint64, lastMessageAt time.Time) (sql.Result, error) {
	return m.withSession(session).IncrementMessageCount(ctx, id, lastMessageAt)
}

func (m *customChatSessionModel) UpdateTitle(ctx context.Context, id uint64, title string) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s 
		set title = ?
		where id = ?
			and deleted_at = 0`,
		m.table)

	result, err := m.conn.ExecCtx(ctx, query, title, id)
	return result, mapDBError(err)
}
