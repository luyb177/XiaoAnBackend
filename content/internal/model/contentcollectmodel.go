package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ContentCollectModel = (*customContentCollectModel)(nil)

type (
	// ContentCollectModel is an interface to be customized, add more methods here,
	// and implement the added methods in customContentCollectModel.
	ContentCollectModel interface {
		contentCollectModel
		withSession(session sqlx.Session) ContentCollectModel
		FindOneByUserIDTypeTargetID(ctx context.Context, userID uint64, tp string, targetID uint64) (*ContentCollect, error)
		FindOneByUserIDTypeTargetIDWithSession(ctx context.Context, session sqlx.Session, userID uint64, tp string, targetID uint64) (*ContentCollect, error)
		Upsert(ctx context.Context, data *ContentCollect) (sql.Result, error)
		UpsertWithSession(ctx context.Context, session sqlx.Session, data *ContentCollect) (sql.Result, error)
		MarkCollectAsCounted(ctx context.Context, id uint64) (sql.Result, error)
		MarkCollectAsCountedWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
		UnmarkCollectAsCounted(ctx context.Context, id uint64) (sql.Result, error)
		UnmarkCollectAsCountedWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
		SoftDelete(ctx context.Context, id, deletedAt uint64) (sql.Result, error)
		SoftDeleteWithSession(ctx context.Context, session sqlx.Session, id, deletedAt uint64) (sql.Result, error)
	}

	customContentCollectModel struct {
		*defaultContentCollectModel
	}
)

// NewContentCollectModel returns a model for the database table.
func NewContentCollectModel(conn sqlx.SqlConn) ContentCollectModel {
	return &customContentCollectModel{
		defaultContentCollectModel: newContentCollectModel(conn),
	}
}

func (m *customContentCollectModel) withSession(session sqlx.Session) ContentCollectModel {
	return NewContentCollectModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customContentCollectModel) FindOneByUserIDTypeTargetID(ctx context.Context, userID uint64, tp string, targetID uint64) (*ContentCollect, error) {
	//nolint:staticcheck // QF1008: go-zero embedding style retained intentionally
	res, err := m.defaultContentCollectModel.FindOneByUserIdTypeTargetId(ctx, userID, tp, targetID)
	return res, mapDBError(err)
}

func (m *customContentCollectModel) FindOneByUserIDTypeTargetIDWithSession(ctx context.Context, session sqlx.Session, userID uint64, tp string, targetID uint64) (*ContentCollect, error) {
	return m.withSession(session).FindOneByUserIDTypeTargetID(ctx, userID, tp, targetID)
}

func (m *customContentCollectModel) Upsert(ctx context.Context, data *ContentCollect) (sql.Result, error) {
	query := fmt.Sprintf(`
		INSERT INTO %s (%s)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			deleted_at = 0`,
		m.table,
		contentCollectRowsExpectAutoSet,
	)

	result, err := m.conn.ExecCtx(ctx, query,
		data.Type,
		data.TargetId,
		data.UserId,
		data.IsCounted,
		data.DeletedAt,
	)
	return result, mapDBError(err)
}

func (m *customContentCollectModel) UpsertWithSession(ctx context.Context, session sqlx.Session, data *ContentCollect) (sql.Result, error) {
	return m.withSession(session).Upsert(ctx, data)
}

func (m *customContentCollectModel) MarkCollectAsCounted(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set is_counted = 1
		where id = ? and is_counted = 0 and deleted_at = 0`,
		m.table)
	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}
func (m *customContentCollectModel) MarkCollectAsCountedWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).MarkCollectAsCounted(ctx, id)
}

func (m *customContentCollectModel) UnmarkCollectAsCounted(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set is_counted = 0
		where id = ? and is_counted = 1 and deleted_at = 0`,
		m.table)
	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customContentCollectModel) UnmarkCollectAsCountedWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).UnmarkCollectAsCounted(ctx, id)
}

func (m *customContentCollectModel) SoftDelete(ctx context.Context, id, deletedAt uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set deleted_at = ?
		where id = ? and deleted_at = 0`,
		m.table)

	result, err := m.conn.ExecCtx(ctx, query, deletedAt, id)
	return result, mapDBError(err)
}
func (m *customContentCollectModel) SoftDeleteWithSession(ctx context.Context, session sqlx.Session, id, deletedAt uint64) (sql.Result, error) {
	return m.withSession(session).SoftDelete(ctx, id, deletedAt)
}
