package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ComicPageModel = (*customComicPageModel)(nil)

type (
	// ComicPageModel is an interface to be customized, add more methods here,
	// and implement the added methods in customComicPageModel.
	ComicPageModel interface {
		comicPageModel
		withSession(session sqlx.Session) ComicPageModel
		InsertBatch(ctx context.Context, list []*ComicPage) error
		InsertBatchWithSession(ctx context.Context, session sqlx.Session, list []*ComicPage) error
		FindManyByChapterIDOrderByPageNo(ctx context.Context, chapterID uint64, offset, pageSize int64) ([]*ComicPage, error)
		SoftDeleteByChapterID(ctx context.Context, chapterID uint64, deletedAt uint64) error
		SoftDeleteByChapterIDWithSession(ctx context.Context, session sqlx.Session, chapterID uint64, deletedAt uint64) error
		SoftDeleteAllByChapterIDs(ctx context.Context, chapterIDs []uint64, deletedAt uint64) error
		SoftDeleteAllByChapterIDsWithSession(ctx context.Context, session sqlx.Session, chapterIDs []uint64, deletedAt uint64) error
	}

	customComicPageModel struct {
		*defaultComicPageModel
	}
)

// NewComicPageModel returns a model for the database table.
func NewComicPageModel(conn sqlx.SqlConn) ComicPageModel {
	return &customComicPageModel{
		defaultComicPageModel: newComicPageModel(conn),
	}
}

func (m *customComicPageModel) withSession(session sqlx.Session) ComicPageModel {
	return NewComicPageModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customComicPageModel) InsertBatch(ctx context.Context, list []*ComicPage) error {
	if len(list) == 0 {
		return nil
	}

	// 构造 values
	valuePlaceholders := make([]string, 0, len(list))
	args := make([]interface{}, 0, len(list)*4)

	for _, page := range list {
		valuePlaceholders = append(valuePlaceholders, "(?,?,?,?)")
		args = append(args, page.ChapterId, page.PageNo, page.Url, page.DeletedAt)
	}
	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES %s`,
		m.table,
		comicPageRowsExpectAutoSet,
		strings.Join(valuePlaceholders, ","),
	)
	_, err := m.conn.ExecCtx(ctx, query, args...)
	return mapDBError(err)
}

func (m *customComicPageModel) InsertBatchWithSession(ctx context.Context, session sqlx.Session, list []*ComicPage) error {
	return m.withSession(session).InsertBatch(ctx, list)
}

func (m *customComicPageModel) FindManyByChapterIDOrderByPageNo(ctx context.Context, chapterID uint64, offset, pageSize int64) ([]*ComicPage, error) {
	query := fmt.Sprintf(
		"select %s from %s where `chapter_id` = ? and `deleted_at` = 0 order by `page_no` asc limit ? offset ?",
		comicPageRows,
		m.table,
	)

	var resp []*ComicPage
	err := m.conn.QueryRowsCtx(ctx, &resp, query, chapterID, pageSize, offset)
	return resp, mapDBError(err)
}

func (m *customComicPageModel) SoftDeleteByChapterID(ctx context.Context, chapterID uint64, deletedAt uint64) error {
	query := fmt.Sprintf(
		"update %s set `deleted_at` = ? where `chapter_id` = ? and `deleted_at` = 0",
		m.table,
	)
	_, err := m.conn.ExecCtx(ctx, query, deletedAt, chapterID)
	return mapDBError(err)
}

func (m *customComicPageModel) SoftDeleteByChapterIDWithSession(ctx context.Context, session sqlx.Session, chapterID uint64, deletedAt uint64) error {
	return m.withSession(session).SoftDeleteByChapterID(ctx, chapterID, deletedAt)
}

func (m *customComicPageModel) SoftDeleteAllByChapterIDs(ctx context.Context, chapterIDs []uint64, deletedAt uint64) error {
	if len(chapterIDs) == 0 {
		return nil
	}

	idPlaceholders := make([]string, len(chapterIDs))
	args := make([]interface{}, 0, len(chapterIDs)+1) // 1个公共参数 + N 个 id
	args = append(args, deletedAt)
	for i, id := range chapterIDs {
		idPlaceholders[i] = "?"
		args = append(args, id)
	}
	query := fmt.Sprintf(
		"update %s set `deleted_at` = ? where `chapter_id` in (%s) and `deleted_at` = 0",
		m.table,
		strings.Join(idPlaceholders, ","),
	)

	_, err := m.conn.ExecCtx(ctx, query, args...)
	return mapDBError(err)
}

func (m *customComicPageModel) SoftDeleteAllByChapterIDsWithSession(ctx context.Context, session sqlx.Session, chapterIDs []uint64, deletedAt uint64) error {
	return m.withSession(session).SoftDeleteAllByChapterIDs(ctx, chapterIDs, deletedAt)
}
