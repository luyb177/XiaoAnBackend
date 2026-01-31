package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CommentModel = (*customCommentModel)(nil)

const (
	DefaultOffset = 0
	DefaultLimit  = 3
)

type (
	// CommentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCommentModel.
	CommentModel interface {
		commentModel
		withSession(session sqlx.Session) CommentModel
		FindOneWithNotDeleted(ctx context.Context, id uint64) (*Comment, error)
		FindByTypeAndTargetId(ctx context.Context, tp string, targetId uint64, offset int64, limit int64) (list []*Comment, err error)
		FindOneByTypeAndCommentId(ctx context.Context, tp string, commentId uint64) (*Comment, error)
		FindOneByTypeAndCommentIdWithSession(ctx context.Context, session sqlx.Session, tp string, commentId uint64) (*Comment, error)
		FindOneWithSession(ctx context.Context, session sqlx.Session, id uint64) (*Comment, error)
		FindRootByTypeAndTargetId(ctx context.Context, tp string, targetId uint64, offset int64, limit int64) (list []*Comment, err error)
		FindSubByTypeAndTargetIdAndParentId(ctx context.Context, tp string, targetId uint64, parentId uint64, offset int64, limit int64) (list []*Comment, err error)
		MarkCommentAsCounted(ctx context.Context, commentId uint64) (sql.Result, error)
		MarkCommentAsCountedWithSession(ctx context.Context, session sqlx.Session, commentId uint64) (sql.Result, error)
		UnmarkCommentAsCounted(ctx context.Context, commentId uint64) (sql.Result, error)
		UnmarkCommentAsCountedWithSession(ctx context.Context, session sqlx.Session, commentId uint64) (sql.Result, error)
		UnmarkChildrenCommentAsCounted(ctx context.Context, parentId uint64) (sql.Result, error)
		UnmarkChildrenCommentAsCountedWithSession(ctx context.Context, session sqlx.Session, parentId uint64) (sql.Result, error)
		CountAllByTypeAndTargetId(ctx context.Context, tp string, targetId uint64) (int64, error)
		CountParentByTypeAndTargetId(ctx context.Context, tp string, targetId uint64) (int64, error)
		FindChildByTypeAndTargetId(ctx context.Context, tp string, targetId uint64, parentId uint64, offset int64, limit int64) (list []*Comment, err error)
		CountChildByTypeAndTargetId(ctx context.Context, tp string, targetId uint64, parentId uint64) (int64, error)
		FindDefaultChildByTypeAndTargetId(ctx context.Context, tp string, targetId uint64, parentId uint64) (list []*Comment, err error)
		IncrSubCommentCount(ctx context.Context, parentId uint64) (sql.Result, error)
		IncrSubCommentCountWithSession(ctx context.Context, session sqlx.Session, parentId uint64) (sql.Result, error)
		DecrSubCommentCount(ctx context.Context, parentId uint64) (sql.Result, error)
		DecrSubCommentCountWithSession(ctx context.Context, session sqlx.Session, parentId uint64) (sql.Result, error)
		UserSoftDelete(ctx context.Context, commentId uint64, deletedAt uint64) error
		UserSoftDeleteWithSession(ctx context.Context, session sqlx.Session, commentId uint64, deletedAt uint64) error
		CascadeSoftDeleteChildren(ctx context.Context, parentId uint64, deletedAt uint64) (sql.Result, error)
		CascadeSoftDeleteChildrenWithSession(ctx context.Context, session sqlx.Session, parentId uint64, deletedAt uint64) (sql.Result, error)
		UpdateStatusWithRemark(ctx context.Context, commentId uint64, status uint64, remark string) error
		UpdateStatusWithRemarkWithSession(ctx context.Context, session sqlx.Session, commentId uint64, status uint64, remark string) error
	}

	customCommentModel struct {
		*defaultCommentModel
	}
)

// NewCommentModel returns a model for the database table.
func NewCommentModel(conn sqlx.SqlConn) CommentModel {
	return &customCommentModel{
		defaultCommentModel: newCommentModel(conn),
	}
}

func (m *customCommentModel) withSession(session sqlx.Session) CommentModel {
	return NewCommentModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customCommentModel) FindOneWithNotDeleted(ctx context.Context, id uint64) (*Comment, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? and `deleted_at` = 0 limit 1", commentRows, m.table)

	var resp Comment
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	return &resp, mapDBError(err)
}

func (m *customCommentModel) FindByTypeAndTargetId(ctx context.Context, tp string, targetId uint64, offset int64, limit int64) (list []*Comment, err error) {
	query := fmt.Sprintf("select %s from %s where `type` = ? and `target_id` = ? and parent_id = 0 limit ?, ?", commentRows, m.table)

	var out []*Comment
	err = m.conn.QueryRowsCtx(ctx, &out, query, tp, targetId, offset, limit)
	return out, mapDBError(err)
}

func (m *customCommentModel) FindOneByTypeAndCommentId(ctx context.Context, tp string, commentId uint64) (*Comment, error) {
	query := fmt.Sprintf("select %s from %s where `type` = ? and `id` = ? and `deleted_at` = 0 limit 1", commentRows, m.table)

	var resp Comment
	err := m.conn.QueryRowCtx(ctx, &resp, query, tp, commentId)
	return &resp, mapDBError(err)
}

func (m *customCommentModel) FindOneWithSession(ctx context.Context, session sqlx.Session, id uint64) (*Comment, error) {
	return m.withSession(session).FindOne(ctx, id)
}

func (m *customCommentModel) FindOneByTypeAndCommentIdWithSession(ctx context.Context, session sqlx.Session, tp string, commentId uint64) (*Comment, error) {
	return m.withSession(session).FindOneByTypeAndCommentId(ctx, tp, commentId)
}

func (m *customCommentModel) FindRootByTypeAndTargetId(ctx context.Context, tp string, targetId uint64, offset int64, limit int64) (list []*Comment, err error) {
	query := fmt.Sprintf(`
		select %s from %s 
		where type = ? and target_id = ? and parent_id = 0 and deleted_at = 0 and status = 0
		order by like_count desc, created_at desc
		limit ? offset ?`,
		commentRows,
		m.table,
	)

	var out []*Comment
	err = m.conn.QueryRowsCtx(ctx, &out, query, tp, targetId, limit, offset)
	return out, mapDBError(err)
}

func (m *customCommentModel) FindSubByTypeAndTargetIdAndParentId(ctx context.Context, tp string, targetId uint64, parentId uint64, offset int64, limit int64) (list []*Comment, err error) {
	query := fmt.Sprintf(`
		select %s from %s 
		where type = ? and target_id = ? and parent_id = ? and deleted_at = 0 and status = 0
		order by like_count desc, created_at desc
		limit ? offset ?`,
		commentRows,
		m.table,
	)

	var out []*Comment
	err = m.conn.QueryRowsCtx(ctx, &out, query, tp, targetId, parentId, limit, offset)
	return out, mapDBError(err)
}

func (m *customCommentModel) MarkCommentAsCounted(ctx context.Context, commentId uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
        UPDATE %s
        SET is_counted = 1
        WHERE id = ? AND is_counted = 0 AND deleted_at = 0
    `, m.table)

	result, err := m.conn.ExecCtx(ctx, query, commentId)
	return result, mapDBError(err)
}

func (m *customCommentModel) MarkCommentAsCountedWithSession(ctx context.Context, session sqlx.Session, commentId uint64) (sql.Result, error) {
	return m.withSession(session).MarkCommentAsCounted(ctx, commentId)
}

func (m *customCommentModel) UnmarkCommentAsCounted(ctx context.Context, commentId uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
        UPDATE %s
        SET is_counted = 0
        WHERE id = ? AND is_counted = 1
    `, m.table)

	result, err := m.conn.ExecCtx(ctx, query, commentId)
	return result, mapDBError(err)
}

func (m *customCommentModel) UnmarkCommentAsCountedWithSession(ctx context.Context, session sqlx.Session, commentId uint64) (sql.Result, error) {
	return m.withSession(session).UnmarkCommentAsCounted(ctx, commentId)
}

func (m *customCommentModel) UnmarkChildrenCommentAsCounted(ctx context.Context, parentId uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET is_counted = 0
		WHERE parent_id = ? AND is_counted = 1 AND deleted_at = 0
	`, m.table)

	result, err := m.conn.ExecCtx(ctx, query, parentId)
	return result, mapDBError(err)
}

func (m *customCommentModel) UnmarkChildrenCommentAsCountedWithSession(ctx context.Context, session sqlx.Session, parentId uint64) (sql.Result, error) {
	return m.withSession(session).UnmarkChildrenCommentAsCounted(ctx, parentId)
}

func (m *customCommentModel) CountAllByTypeAndTargetId(ctx context.Context, tp string, targetId uint64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where `type` = ? and `target_id` = ?", m.table)

	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, tp, targetId)
	return count, mapDBError(err)
}

func (m *customCommentModel) CountParentByTypeAndTargetId(ctx context.Context, tp string, targetId uint64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where `type` = ? and `target_id` = ? and parent_id = 0", m.table)

	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, tp, targetId)
	return count, mapDBError(err)
}

func (m *customCommentModel) FindDefaultChildByTypeAndTargetId(ctx context.Context, tp string, targetId uint64, parentId uint64) (list []*Comment, err error) {
	query := fmt.Sprintf("select %s from %s where `type` = ? and `target_id` = ? and parent_id = ? limit %d, %d", commentRows, m.table, DefaultOffset, DefaultLimit)

	var out []*Comment
	err = m.conn.QueryRowsCtx(ctx, &out, query, tp, targetId, parentId)
	return out, mapDBError(err)
}

func (m *customCommentModel) FindChildByTypeAndTargetId(ctx context.Context, tp string, targetId uint64, parentId uint64, offset int64, limit int64) (list []*Comment, err error) {
	query := fmt.Sprintf("select %s from %s where `type` = ? and `target_id` = ? and parent_id = ? limit ?, ?", commentRows, m.table)

	var out []*Comment
	err = m.conn.QueryRowsCtx(ctx, &out, query, tp, targetId, parentId, offset, limit)
	return out, mapDBError(err)
}

func (m *customCommentModel) CountChildByTypeAndTargetId(ctx context.Context, tp string, targetId uint64, parentId uint64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where `type` = ? and `target_id` = ? and parent_id = ?", m.table)

	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, tp, targetId, parentId)
	return count, mapDBError(err)
}

func (m *customCommentModel) IncrSubCommentCount(ctx context.Context, parentId uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET sub_comment_count = sub_comment_count + 1
		WHERE id = ? AND deleted_at = 0 AND parent_id = 0
	`, m.table)

	result, err := m.conn.ExecCtx(ctx, query, parentId)
	return result, mapDBError(err)
}

func (m *customCommentModel) IncrSubCommentCountWithSession(ctx context.Context, session sqlx.Session, parentId uint64) (sql.Result, error) {
	return m.withSession(session).IncrSubCommentCount(ctx, parentId)
}

func (m *customCommentModel) DecrSubCommentCount(ctx context.Context, parentId uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
	UPDATE %s
	SET sub_comment_count = sub_comment_count - 1
	WHERE id = ?
		AND deleted_at = 0
		AND sub_comment_count > 0`,
		m.table)
	result, err := m.conn.ExecCtx(ctx, query, parentId)
	return result, mapDBError(err)
}

func (m *customCommentModel) DecrSubCommentCountWithSession(ctx context.Context, session sqlx.Session, parentId uint64) (sql.Result, error) {
	return m.withSession(session).DecrSubCommentCount(ctx, parentId)
}

// UserSoftDelete 表示用户删除这条评论 未碰 is_counted 字段
func (m *customCommentModel) UserSoftDelete(ctx context.Context, commentId uint64, deletedAt uint64) error {
	query := fmt.Sprintf(`
		update %s
		set deleted_at = ?
		where id = ? and deleted_at = 0`,
		m.table,
	)
	_, err := m.conn.ExecCtx(ctx, query, deletedAt, commentId)
	return mapDBError(err)
}

func (m *customCommentModel) UserSoftDeleteWithSession(ctx context.Context, session sqlx.Session, commentId uint64, deletedAt uint64) error {
	return m.withSession(session).UserSoftDelete(ctx, commentId, deletedAt)
}

// CascadeSoftDeleteChildren 表示级联删除子评论, 未碰 is_counted 字段
func (m *customCommentModel) CascadeSoftDeleteChildren(ctx context.Context, parentId uint64, deletedAt uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
	update %s
	set deleted_at = ?
	where parent_id = ? and deleted_at = 0`,
		m.table)

	result, err := m.conn.ExecCtx(ctx, query, deletedAt, parentId)
	return result, mapDBError(err)
}

func (m *customCommentModel) CascadeSoftDeleteChildrenWithSession(ctx context.Context, session sqlx.Session, parentId uint64, deletedAt uint64) (sql.Result, error) {
	return m.withSession(session).CascadeSoftDeleteChildren(ctx, parentId, deletedAt)
}

func (m *customCommentModel) UpdateStatusWithRemark(ctx context.Context, commentId uint64, status uint64, remark string) error {
	query := fmt.Sprintf("update %s set `status` = ?, `remark` = ? where `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, remark, commentId)
	return mapDBError(err)
}

func (m *customCommentModel) UpdateStatusWithRemarkWithSession(ctx context.Context, session sqlx.Session, commentId uint64, status uint64, remark string) error {
	return m.withSession(session).UpdateStatusWithRemark(ctx, commentId, status, remark)
}
