package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ComicTagModel = (*customComicTagModel)(nil)

type (
	// ComicTagModel is an interface to be customized, add more methods here,
	// and implement the added methods in customComicTagModel.
	ComicTagModel interface {
		comicTagModel
		withSession(session sqlx.Session) ComicTagModel
		FindManyByComicID(ctx context.Context, comicID uint64) ([]*ComicTag, error)
		InsertBatch(ctx context.Context, list []*ComicTag) error
		InsertBatchWithSession(ctx context.Context, session sqlx.Session, list []*ComicTag) error
		SoftDeleteBatchByComicID(ctx context.Context, comicID, deletedAt uint64) error
		SoftDeleteBatchByComicIDWithSession(ctx context.Context, session sqlx.Session, comicID, deletedAt uint64) error
	}

	customComicTagModel struct {
		*defaultComicTagModel
	}
)

// NewComicTagModel returns a model for the database table.
func NewComicTagModel(conn sqlx.SqlConn) ComicTagModel {
	return &customComicTagModel{
		defaultComicTagModel: newComicTagModel(conn),
	}
}

func (m *customComicTagModel) withSession(session sqlx.Session) ComicTagModel {
	return NewComicTagModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customComicTagModel) FindManyByComicID(ctx context.Context, comicID uint64) ([]*ComicTag, error) {
	query := fmt.Sprintf(
		"select %s from %s where `comic_id` = ? and `deleted_at` = 0",
		comicTagRows,
		m.table,
	)
	var resp []*ComicTag
	err := m.conn.QueryRowsCtx(ctx, &resp, query, comicID)
	return resp, mapDBError(err)
}

func (m *customComicTagModel) InsertBatch(ctx context.Context, list []*ComicTag) error {
	if len(list) == 0 {
		return nil
	}

	// 构造 values
	valuePlaceholders := make([]string, 0, len(list))
	args := make([]interface{}, 0, len(list)*3)

	for _, tag := range list {
		valuePlaceholders = append(valuePlaceholders, "(?,?,?)")
		args = append(args, tag.ComicId, tag.Tag, tag.DeletedAt)
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES %s`,
		m.table,
		comicTagRowsExpectAutoSet,
		strings.Join(valuePlaceholders, ","),
	)

	_, err := m.conn.ExecCtx(ctx, query, args...)
	return mapDBError(err)
}

func (m *customComicTagModel) InsertBatchWithSession(ctx context.Context, session sqlx.Session, list []*ComicTag) error {
	return m.withSession(session).InsertBatch(ctx, list)
}

func (m *customComicTagModel) SoftDeleteBatchByComicID(ctx context.Context, comicID, deletedAt uint64) error {
	query := fmt.Sprintf(
		"update %s set `deleted_at` = ? where `comic_id` = ? and `deleted_at` = 0",
		m.table,
	)
	_, err := m.conn.ExecCtx(ctx, query, deletedAt, comicID)
	return mapDBError(err)
}

func (m *customComicTagModel) SoftDeleteBatchByComicIDWithSession(ctx context.Context, session sqlx.Session, comicID, deletedAt uint64) error {
	return m.withSession(session).SoftDeleteBatchByComicID(ctx, comicID, deletedAt)
}
