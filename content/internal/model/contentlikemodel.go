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
		FindOneByUserIdTypeTargetId(ctx context.Context, userId uint64, tp string, targetId uint64) (*ContentLike, error)
		FindOneByUserIdTypeTargetIdWithSession(ctx context.Context, session sqlx.Session, userId uint64, tp string, targetId uint64) (*ContentLike, error)
		Upsert(ctx context.Context, data *ContentLike) (sql.Result, error)
		UpsertWithSession(ctx context.Context, session sqlx.Session, data *ContentLike) (sql.Result, error)
		SoftDelete(ctx context.Context, id uint64, deletedAt uint64) (sql.Result, error)
		SoftDeleteWithSession(ctx context.Context, session sqlx.Session, id uint64, deletedAt uint64) (sql.Result, error)
		SoftDeleteByTypeTargetId(ctx context.Context, tp string, targetId uint64, deletedAt uint64) (sql.Result, error)
		SoftDeleteByTypeTargetIdWithSession(ctx context.Context, session sqlx.Session, tp string, targetId uint64, deletedAt uint64) (sql.Result, error)
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

func (m *customContentLikeModel) FindOneByUserIdTypeTargetId(ctx context.Context, userId uint64, tp string, targetId uint64) (*ContentLike, error) {
	res, err := m.defaultContentLikeModel.FindOneByUserIdTypeTargetId(ctx, userId, tp, targetId)
	return res, mapDBError(err)
}

func (m *customContentLikeModel) FindOneByUserIdTypeTargetIdWithSession(ctx context.Context, session sqlx.Session, userId uint64, tp string, targetId uint64) (*ContentLike, error) {
	return m.withSession(session).FindOneByUserIdTypeTargetId(ctx, userId, tp, targetId)
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

func (m *customContentLikeModel) SoftDelete(ctx context.Context, id uint64, deletedAt uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set deleted_at = ?
		where id = ? and deleted_at = 0`,
		m.table)

	result, err := m.conn.ExecCtx(ctx, query, deletedAt, id)
	return result, mapDBError(err)
}

func (m *customContentLikeModel) SoftDeleteWithSession(ctx context.Context, session sqlx.Session, id uint64, deletedAt uint64) (sql.Result, error) {
	return m.withSession(session).SoftDelete(ctx, id, deletedAt)
}

func (m *customContentLikeModel) SoftDeleteByTypeTargetId(ctx context.Context, tp string, targetId uint64, deletedAt uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set deleted_at = ?
		where type = ? and target_id = ? and deleted_at = 0`,
		m.table,
	)

	result, err := m.conn.ExecCtx(ctx, query, deletedAt, tp, targetId)
	return result, mapDBError(err)

}

func (m *customContentLikeModel) SoftDeleteByTypeTargetIdWithSession(ctx context.Context, session sqlx.Session, tp string, targetId uint64, deletedAt uint64) (sql.Result, error) {
	return m.withSession(session).SoftDeleteByTypeTargetId(ctx, tp, targetId, deletedAt)
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
