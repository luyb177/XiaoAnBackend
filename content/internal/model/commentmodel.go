package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CommentModel = (*customCommentModel)(nil)

type (
	// CommentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCommentModel.
	CommentModel interface {
		commentModel
		withSession(session sqlx.Session) CommentModel
		FindOneWithNotDeleted(ctx context.Context, id uint64) (*Comment, error)
		FindOneWithSession(ctx context.Context, session sqlx.Session, id uint64) (*Comment, error)
		FindRootByTypeAndTargetID(ctx context.Context, tp string, targetID uint64, offset int64, limit int64) (list []*Comment, err error)
		FindSubByTypeAndTargetIDAndParentID(ctx context.Context, tp string, targetID, parentID uint64, offset, limit int64) (list []*Comment, err error)
		MarkCommentAsCounted(ctx context.Context, commentID uint64) (sql.Result, error)
		MarkCommentAsCountedWithSession(ctx context.Context, session sqlx.Session, commentID uint64) (sql.Result, error)
		UnmarkCommentAsCounted(ctx context.Context, commentID uint64) (sql.Result, error)
		UnmarkCommentAsCountedWithSession(ctx context.Context, session sqlx.Session, commentID uint64) (sql.Result, error)
		UnmarkChildrenCommentAsCounted(ctx context.Context, parentID uint64) (sql.Result, error)
		UnmarkChildrenCommentAsCountedWithSession(ctx context.Context, session sqlx.Session, parentID uint64) (sql.Result, error)
		IncrSubCommentCount(ctx context.Context, parentID uint64) (sql.Result, error)
		IncrSubCommentCountWithSession(ctx context.Context, session sqlx.Session, parentID uint64) (sql.Result, error)
		DecrSubCommentCount(ctx context.Context, parentID uint64) (sql.Result, error)
		DecrSubCommentCountWithSession(ctx context.Context, session sqlx.Session, parentID uint64) (sql.Result, error)
		IncrLikeCount(ctx context.Context, id uint64) (sql.Result, error)
		IncrLikeCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
		DecrLikeCount(ctx context.Context, id uint64) (sql.Result, error)
		DecrLikeCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
		UserSoftDelete(ctx context.Context, commentID uint64, deletedAt uint64) error
		UserSoftDeleteWithSession(ctx context.Context, session sqlx.Session, commentID uint64, deletedAt uint64) error
		CascadeSoftDeleteChildren(ctx context.Context, parentID uint64, deletedAt uint64) (sql.Result, error)
		CascadeSoftDeleteChildrenWithSession(ctx context.Context, session sqlx.Session, parentID uint64, deletedAt uint64) (sql.Result, error)
		SoftDeleteByTypeAndTargetID(ctx context.Context, tp string, targetID, deletedAt uint64) (sql.Result, error)
		SoftDeleteByTypeAndTargetIDWithSession(ctx context.Context, session sqlx.Session, tp string, targetID, deletedAt uint64) (sql.Result, error)
		UpdateNicknameByUserId(ctx context.Context, userId uint64, nickname string) error
		UpdateAvatarByUserId(ctx context.Context, userId uint64, avatar string) error
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

func (m *customCommentModel) FindOneWithSession(ctx context.Context, session sqlx.Session, id uint64) (*Comment, error) {
	return m.withSession(session).FindOne(ctx, id)
}

func (m *customCommentModel) FindOneWithNotDeleted(ctx context.Context, id uint64) (*Comment, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? and `deleted_at` = 0 limit 1", commentRows, m.table)

	var resp Comment
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	return &resp, mapDBError(err)
}

func (m *customCommentModel) FindRootByTypeAndTargetID(ctx context.Context, tp string, targetID uint64, offset, limit int64) (list []*Comment, err error) {
	query := fmt.Sprintf(`
		select %s from %s 
		where type = ? and target_id = ? and parent_id = 0 and deleted_at = 0 and status = 0
		order by like_count desc, created_at desc
		limit ? offset ?`,
		commentRows,
		m.table,
	)

	var out []*Comment
	err = m.conn.QueryRowsCtx(ctx, &out, query, tp, targetID, limit, offset)
	return out, mapDBError(err)
}

func (m *customCommentModel) FindSubByTypeAndTargetIDAndParentID(ctx context.Context, tp string, targetID, parentID uint64, offset, limit int64) (list []*Comment, err error) {
	query := fmt.Sprintf(`
		select %s from %s 
		where type = ? and target_id = ? and parent_id = ? and deleted_at = 0 and status = 0
		order by like_count desc, created_at desc
		limit ? offset ?`,
		commentRows,
		m.table,
	)

	var out []*Comment
	err = m.conn.QueryRowsCtx(ctx, &out, query, tp, targetID, parentID, limit, offset)
	return out, mapDBError(err)
}

func (m *customCommentModel) MarkCommentAsCounted(ctx context.Context, commentID uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
        UPDATE %s
        SET is_counted = 1
        WHERE id = ? AND is_counted = 0 AND deleted_at = 0
    `, m.table)

	result, err := m.conn.ExecCtx(ctx, query, commentID)
	return result, mapDBError(err)
}

func (m *customCommentModel) MarkCommentAsCountedWithSession(ctx context.Context, session sqlx.Session, commentID uint64) (sql.Result, error) {
	return m.withSession(session).MarkCommentAsCounted(ctx, commentID)
}

func (m *customCommentModel) UnmarkCommentAsCounted(ctx context.Context, commentID uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
        UPDATE %s
        SET is_counted = 0
        WHERE id = ? AND is_counted = 1
    `, m.table)

	result, err := m.conn.ExecCtx(ctx, query, commentID)
	return result, mapDBError(err)
}

func (m *customCommentModel) UnmarkCommentAsCountedWithSession(ctx context.Context, session sqlx.Session, commentID uint64) (sql.Result, error) {
	return m.withSession(session).UnmarkCommentAsCounted(ctx, commentID)
}

func (m *customCommentModel) UnmarkChildrenCommentAsCounted(ctx context.Context, parentID uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET is_counted = 0
		WHERE parent_id = ? AND is_counted = 1 AND deleted_at = 0
	`, m.table)

	result, err := m.conn.ExecCtx(ctx, query, parentID)
	return result, mapDBError(err)
}

func (m *customCommentModel) UnmarkChildrenCommentAsCountedWithSession(ctx context.Context, session sqlx.Session, parentID uint64) (sql.Result, error) {
	return m.withSession(session).UnmarkChildrenCommentAsCounted(ctx, parentID)
}

func (m *customCommentModel) IncrSubCommentCount(ctx context.Context, parentID uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET sub_comment_count = sub_comment_count + 1
		WHERE id = ? AND deleted_at = 0 AND parent_id = 0
	`, m.table)

	result, err := m.conn.ExecCtx(ctx, query, parentID)
	return result, mapDBError(err)
}

func (m *customCommentModel) IncrSubCommentCountWithSession(ctx context.Context, session sqlx.Session, parentID uint64) (sql.Result, error) {
	return m.withSession(session).IncrSubCommentCount(ctx, parentID)
}

func (m *customCommentModel) DecrSubCommentCount(ctx context.Context, parentID uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
	UPDATE %s
	SET sub_comment_count = sub_comment_count - 1
	WHERE id = ?
		AND deleted_at = 0
		AND sub_comment_count > 0`,
		m.table)
	result, err := m.conn.ExecCtx(ctx, query, parentID)
	return result, mapDBError(err)
}

func (m *customCommentModel) DecrSubCommentCountWithSession(ctx context.Context, session sqlx.Session, parentID uint64) (sql.Result, error) {
	return m.withSession(session).DecrSubCommentCount(ctx, parentID)
}

func (m *customCommentModel) IncrLikeCount(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET like_count = like_count + 1
		WHERE id = ? AND deleted_at = 0
	`, m.table)

	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customCommentModel) IncrLikeCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).IncrLikeCount(ctx, id)
}

func (m *customCommentModel) DecrLikeCount(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set like_count = like_count - 1
		where id = ? and like_count > 0 and deleted_at = 0`,
		m.table,
	)

	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customCommentModel) DecrLikeCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).DecrLikeCount(ctx, id)
}

// UserSoftDelete 表示用户删除这条评论 未碰 is_counted 字段
func (m *customCommentModel) UserSoftDelete(ctx context.Context, commentID, deletedAt uint64) error {
	query := fmt.Sprintf(`
		update %s
		set deleted_at = ?
		where id = ? and deleted_at = 0`,
		m.table,
	)
	_, err := m.conn.ExecCtx(ctx, query, deletedAt, commentID)
	return mapDBError(err)
}

func (m *customCommentModel) UserSoftDeleteWithSession(ctx context.Context, session sqlx.Session, commentID, deletedAt uint64) error {
	return m.withSession(session).UserSoftDelete(ctx, commentID, deletedAt)
}

// CascadeSoftDeleteChildren 表示级联删除子评论, 未碰 is_counted 字段
func (m *customCommentModel) CascadeSoftDeleteChildren(ctx context.Context, parentID, deletedAt uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
	update %s
	set deleted_at = ?
	where parent_id = ? and deleted_at = 0`,
		m.table)

	result, err := m.conn.ExecCtx(ctx, query, deletedAt, parentID)
	return result, mapDBError(err)
}

func (m *customCommentModel) CascadeSoftDeleteChildrenWithSession(ctx context.Context, session sqlx.Session, parentID, deletedAt uint64) (sql.Result, error) {
	return m.withSession(session).CascadeSoftDeleteChildren(ctx, parentID, deletedAt)
}

func (m *customCommentModel) SoftDeleteByTypeAndTargetID(ctx context.Context, tp string, targetID, deletedAt uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
	update %s
	set deleted_at = ?
	where type = ? and target_id = ? and deleted_at = 0`,
		m.table)

	result, err := m.conn.ExecCtx(ctx, query, deletedAt, tp, targetID)
	return result, mapDBError(err)
}

func (m *customCommentModel) SoftDeleteByTypeAndTargetIDWithSession(ctx context.Context, session sqlx.Session, tp string, targetID, deletedAt uint64) (sql.Result, error) {
	return m.withSession(session).SoftDeleteByTypeAndTargetID(ctx, tp, targetID, deletedAt)
}

func (m *customCommentModel) UpdateNicknameByUserId(ctx context.Context, userId uint64, nickname string) error {
	query := fmt.Sprintf("update %s set `nickname` = ? where `user_id` = ? and `deleted_at` = 0", m.table)
	_, err := m.conn.ExecCtx(ctx, query, nickname, userId)
	return mapDBError(err)
}

func (m *customCommentModel) UpdateAvatarByUserId(ctx context.Context, userId uint64, avatar string) error {
	query := fmt.Sprintf("update %s set `avatar` = ? where `user_id` = ? and `deleted_at` = 0", m.table)
	_, err := m.conn.ExecCtx(ctx, query, avatar, userId)
	return mapDBError(err)
}
