package model

import (
	"context"
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
		FindByTypeAndTargetId(ctx context.Context, tp string, targetId uint64, offset int64, limit int64) (list []*Comment, err error)
		CountAllByTypeAndTargetId(ctx context.Context, tp string, targetId uint64) (int64, error)
		CountParentByTypeAndTargetId(ctx context.Context, tp string, targetId uint64) (int64, error)
		FindChildByTypeAndTargetId(ctx context.Context, tp string, targetId uint64, parentId uint64, offset int64, limit int64) (list []*Comment, err error)
		CountChildByTypeAndTargetId(ctx context.Context, tp string, targetId uint64, parentId uint64) (int64, error)
		FindDefaultChildByTypeAndTargetId(ctx context.Context, tp string, targetId uint64, parentId uint64) (list []*Comment, err error)
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

func (m *customCommentModel) FindByTypeAndTargetId(ctx context.Context, tp string, targetId uint64, offset int64, limit int64) (list []*Comment, err error) {
	query := fmt.Sprintf("select %s from %s where `type` = ? and `target_id` = ? and parent_id = 0 limit ?, ?", commentRows, m.table)

	var out []*Comment
	err = m.conn.QueryRowsCtx(ctx, &out, query, tp, targetId, offset, limit)
	return out, err
}

func (m *customCommentModel) CountAllByTypeAndTargetId(ctx context.Context, tp string, targetId uint64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where `type` = ? and `target_id` = ?", m.table)

	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, tp, targetId)
	return count, err
}

func (m *customCommentModel) CountParentByTypeAndTargetId(ctx context.Context, tp string, targetId uint64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where `type` = ? and `target_id` = ? and parent_id = 0", m.table)

	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, tp, targetId)
	return count, err
}

func (m *customCommentModel) FindDefaultChildByTypeAndTargetId(ctx context.Context, tp string, targetId uint64, parentId uint64) (list []*Comment, err error) {
	query := fmt.Sprintf("select %s from %s where `type` = ? and `target_id` = ? and parent_id = ? limit %d, %d", commentRows, m.table, DefaultOffset, DefaultLimit)

	var out []*Comment
	err = m.conn.QueryRowsCtx(ctx, &out, query, tp, targetId, parentId)
	return out, err
}

func (m *customCommentModel) FindChildByTypeAndTargetId(ctx context.Context, tp string, targetId uint64, parentId uint64, offset int64, limit int64) (list []*Comment, err error) {
	query := fmt.Sprintf("select %s from %s where `type` = ? and `target_id` = ? and parent_id = ? limit ?, ?", commentRows, m.table)

	var out []*Comment
	err = m.conn.QueryRowsCtx(ctx, &out, query, tp, targetId, parentId, offset, limit)
	return out, err
}

func (m *customCommentModel) CountChildByTypeAndTargetId(ctx context.Context, tp string, targetId uint64, parentId uint64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where `type` = ? and `target_id` = ? and parent_id = ?", m.table)

	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, tp, targetId, parentId)
	return count, err
}
