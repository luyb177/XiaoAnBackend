package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ContentLikeModel = (*customContentLikeModel)(nil)

type (
	// ContentLikeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customContentLikeModel.
	ContentLikeModel interface {
		contentLikeModel
		withSession(session sqlx.Session) ContentLikeModel
		FindOneByUserIDTypeTargetID(ctx context.Context, userID uint64, tp string, targetID uint64) (*ContentLike, error)
		FindOneByUserIDTypeTargetIDWithSession(ctx context.Context, session sqlx.Session, userID uint64, tp string, targetID uint64) (*ContentLike, error)
		Upsert(ctx context.Context, data *ContentLike) (sql.Result, error)
		UpsertWithSession(ctx context.Context, session sqlx.Session, data *ContentLike) (sql.Result, error)
		SoftDelete(ctx context.Context, id, deletedAt uint64) (sql.Result, error)
		SoftDeleteWithSession(ctx context.Context, session sqlx.Session, id, deletedAt uint64) (sql.Result, error)
		SoftDeleteByTypeTargetID(ctx context.Context, tp string, targetID, deletedAt uint64) (sql.Result, error)
		SoftDeleteByTypeTargetIDWithSession(ctx context.Context, session sqlx.Session, tp string, targetID uint64, deletedAt uint64) (sql.Result, error)
		MarkLikeAsCounted(ctx context.Context, id uint64) (sql.Result, error)
		MarkLikeAsCountedWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
		UnmarkLikeAsCounted(ctx context.Context, id uint64) (sql.Result, error)
		UnmarkLikeAsCountedWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
	}

	customContentLikeModel struct {
		*defaultContentLikeModel
	}
)

// NewContentLikeModel returns a model for the database table.
func NewContentLikeModel(conn sqlx.SqlConn) ContentLikeModel {
	return &customContentLikeModel{
		defaultContentLikeModel: newContentLikeModel(conn),
	}
}

func (m *customContentLikeModel) withSession(session sqlx.Session) ContentLikeModel {
	return NewContentLikeModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customContentLikeModel) FindOneByUserIDTypeTargetID(ctx context.Context, userID uint64, tp string, targetID uint64) (*ContentLike, error) {
	//nolint:staticcheck // QF1008: go-zero embedding style retained intentionally
	res, err := m.defaultContentLikeModel.FindOneByUserIdTypeTargetId(ctx, userID, tp, targetID)
	return res, mapDBError(err)
}

func (m *customContentLikeModel) FindOneByUserIDTypeTargetIDWithSession(ctx context.Context, session sqlx.Session, userID uint64, tp string, targetID uint64) (*ContentLike, error) {
	return m.withSession(session).FindOneByUserIdTypeTargetId(ctx, userID, tp, targetID)
}

// Upsert -> 对象不存在 -> 插入
//
//	-> 对象存在 -> 更新 deleted_at = 0
func (m *customContentLikeModel) Upsert(ctx context.Context, data *ContentLike) (sql.Result, error) {
	query := fmt.Sprintf(`
		INSERT INTO %s (%s) 
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE 
		    deleted_at = 0`,
		m.table,
		contentLikeRowsExpectAutoSet,
	)

	result, err := m.conn.ExecCtx(ctx, query, data.Type, data.TargetId, data.UserId, data.IsCounted, data.DeletedAt)
	return result, mapDBError(err)
}

func (m *customContentLikeModel) UpsertWithSession(ctx context.Context, session sqlx.Session, data *ContentLike) (sql.Result, error) {
	return m.withSession(session).Upsert(ctx, data)
}

func (m *customContentLikeModel) SoftDelete(ctx context.Context, id, deletedAt uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set deleted_at = ?
		where id = ? and deleted_at = 0`,
		m.table)

	result, err := m.conn.ExecCtx(ctx, query, deletedAt, id)
	return result, mapDBError(err)
}

func (m *customContentLikeModel) SoftDeleteWithSession(ctx context.Context, session sqlx.Session, id, deletedAt uint64) (sql.Result, error) {
	return m.withSession(session).SoftDelete(ctx, id, deletedAt)
}

func (m *customContentLikeModel) SoftDeleteByTypeTargetID(ctx context.Context, tp string, targetID, deletedAt uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set deleted_at = ?
		where type = ? and target_id = ? and deleted_at = 0`,
		m.table,
	)

	result, err := m.conn.ExecCtx(ctx, query, deletedAt, tp, targetID)
	return result, mapDBError(err)

}

func (m *customContentLikeModel) SoftDeleteByTypeTargetIDWithSession(ctx context.Context, session sqlx.Session, tp string, targetID, deletedAt uint64) (sql.Result, error) {
	return m.withSession(session).SoftDeleteByTypeTargetID(ctx, tp, targetID, deletedAt)
}

func (m *customContentLikeModel) MarkLikeAsCounted(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET is_counted = 1
		WHERE id = ? AND is_counted = 0 AND deleted_at = 0
	`, m.table)
	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customContentLikeModel) MarkLikeAsCountedWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).MarkLikeAsCounted(ctx, id)
}

func (m *customContentLikeModel) UnmarkLikeAsCounted(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET is_counted=0
		WHERE id = ? AND is_counted = 1 AND deleted_at = 0`,
		m.table,
	)

	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customContentLikeModel) UnmarkLikeAsCountedWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).UnmarkLikeAsCounted(ctx, id)
}
