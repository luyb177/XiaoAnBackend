package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ComicChapterModel = (*customComicChapterModel)(nil)

type (
	// ComicChapterModel is an interface to be customized, add more methods here,
	// and implement the added methods in customComicChapterModel.
	ComicChapterModel interface {
		comicChapterModel
		withSession(session sqlx.Session) ComicChapterModel
		Insert(ctx context.Context, data *ComicChapter) (sql.Result, error)
		InsertWithSession(ctx context.Context, session sqlx.Session, data *ComicChapter) (sql.Result, error)
		FindOneWithNotDelete(ctx context.Context, id uint64) (*ComicChapter, error)
		FindOneByComicIDAndChapterNo(ctx context.Context, comicID uint64, chapterNo int64) (*ComicChapter, error)
		FindOneByComicIDAndChapterID(ctx context.Context, comicID uint64, chapterID uint64) (*ComicChapter, error)
		FindManyByComicIDOrderByChapterNo(ctx context.Context, comicID uint64, offset, pageSize int64) ([]*ComicChapter, error)
		FindAllByComicID(ctx context.Context, comicID uint64) ([]*ComicChapter, error)
		FindAllByComicIDWithSession(ctx context.Context, session sqlx.Session, comicID uint64) ([]*ComicChapter, error)
		UpdateRelationStatus(ctx context.Context, id uint64, relationStatus int64) error
		UpdateRelationStatusWithSession(ctx context.Context, session sqlx.Session, id uint64, relationStatus int64) error
		SoftDelete(ctx context.Context, id uint64, deletedAt uint64, modifier sql.NullInt64) error
		SoftDeleteByIDs(ctx context.Context, ids []uint64, deletedAt uint64, modifier sql.NullInt64) error
		SoftDeleteByIDsWithSession(ctx context.Context, session sqlx.Session, ids []uint64, deletedAt uint64, modifier sql.NullInt64) error
	}

	customComicChapterModel struct {
		*defaultComicChapterModel
	}
)

// NewComicChapterModel returns a model for the database table.
func NewComicChapterModel(conn sqlx.SqlConn) ComicChapterModel {
	return &customComicChapterModel{
		defaultComicChapterModel: newComicChapterModel(conn),
	}
}

func (m *customComicChapterModel) withSession(session sqlx.Session) ComicChapterModel {
	return NewComicChapterModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customComicChapterModel) Insert(ctx context.Context, data *ComicChapter) (sql.Result, error) {
	ret, err := m.defaultComicChapterModel.Insert(ctx, data)
	return ret, mapDBError(err)
}

func (m *customComicChapterModel) InsertWithSession(ctx context.Context, session sqlx.Session, data *ComicChapter) (sql.Result, error) {
	return m.withSession(session).Insert(ctx, data)
}

func (m *customComicChapterModel) FindOneWithNotDelete(ctx context.Context, id uint64) (*ComicChapter, error) {
	query := fmt.Sprintf(
		"select %s from %s where `id` = ? and `deleted_at` = 0 limit 1",
		comicChapterRows,
		m.table,
	)

	var resp ComicChapter
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	return &resp, mapDBError(err)
}

func (m *customComicChapterModel) FindOneByComicIDAndChapterNo(ctx context.Context, comicID uint64, chapterNo int64) (*ComicChapter, error) {
	query := fmt.Sprintf(
		"select %s from %s where `comic_id` = ? and `chapter_no` = ? and `deleted_at` = 0 limit 1",
		comicChapterRows,
		m.table,
	)

	var resp ComicChapter
	err := m.conn.QueryRowCtx(ctx, &resp, query, comicID, chapterNo)
	return &resp, mapDBError(err)
}

func (m *customComicChapterModel) FindOneByComicIDAndChapterID(ctx context.Context, comicID uint64, chapterID uint64) (*ComicChapter, error) {
	query := fmt.Sprintf(
		"select %s from %s where `comic_id` = ? and `id` = ? and `deleted_at` = 0 limit 1",
		comicChapterRows,
		m.table,
	)

	var resp ComicChapter
	err := m.conn.QueryRowCtx(ctx, &resp, query, comicID, chapterID)
	return &resp, mapDBError(err)
}

func (m *customComicChapterModel) FindManyByComicIDOrderByChapterNo(ctx context.Context, comicID uint64, offset, pageSize int64) ([]*ComicChapter, error) {
	query := fmt.Sprintf(
		"select %s from %s where `comic_id` = ? and `deleted_at` = 0 order by `chapter_no` asc limit ? offset ?",
		comicChapterRows,
		m.table,
	)

	var resp []*ComicChapter
	err := m.conn.QueryRowsCtx(ctx, &resp, query, comicID, pageSize, offset)
	return resp, mapDBError(err)
}

func (m *customComicChapterModel) FindAllByComicID(ctx context.Context, comicID uint64) ([]*ComicChapter, error) {
	query := fmt.Sprintf(
		"select %s from %s where `comic_id` = ? and `deleted_at` = 0",
		comicChapterRows,
		m.table,
	)

	var resp []*ComicChapter
	err := m.conn.QueryRowsCtx(ctx, &resp, query, comicID)
	return resp, mapDBError(err)
}

func (m *customComicChapterModel) FindAllByComicIDWithSession(ctx context.Context, session sqlx.Session, comicID uint64) ([]*ComicChapter, error) {
	return m.withSession(session).FindAllByComicID(ctx, comicID)
}

func (m *customComicChapterModel) UpdateRelationStatus(ctx context.Context, id uint64, relationStatus int64) error {
	query := fmt.Sprintf("update %s set `relation_status` = ? where `id` = ?", m.table)

	_, err := m.conn.ExecCtx(ctx, query, relationStatus, id)
	return mapDBError(err)
}

func (m *customComicChapterModel) UpdateRelationStatusWithSession(ctx context.Context, session sqlx.Session, id uint64, relationStatus int64) error {
	return m.withSession(session).UpdateRelationStatus(ctx, id, relationStatus)
}

func (m *customComicChapterModel) SoftDelete(ctx context.Context, id uint64, deletedAt uint64, modifier sql.NullInt64) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET deleted_at = ?, last_modified_by = ?
		WHERE id = ? AND deleted_at = 0
	`, m.table)

	_, err := m.conn.ExecCtx(ctx, query, deletedAt, modifier, id)
	return mapDBError(err)
}

func (m *customComicChapterModel) SoftDeleteByIDs(ctx context.Context, ids []uint64, deletedAt uint64, modifier sql.NullInt64) error {
	if len(ids) == 0 {
		return nil
	}

	idPlaceholders := make([]string, len(ids))
	args := make([]interface{}, 0, 2+len(ids)) // 2个公共参数 + N 个 id
	args = append(args, deletedAt, modifier)
	for i, id := range ids {
		idPlaceholders[i] = "?"
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		UPDATE %s
		SET deleted_at = ?, last_modified_by = ?
		WHERE id IN (%s) AND deleted_at = 0`,
		m.table,
		strings.Join(idPlaceholders, ","),
	)

	// 注意参数顺序：先公共参数，再 id 列表
	_, err := m.conn.ExecCtx(ctx, query, args...)
	return mapDBError(err)
}

func (m *customComicChapterModel) SoftDeleteByIDsWithSession(ctx context.Context, session sqlx.Session, ids []uint64, deletedAt uint64, modifier sql.NullInt64) error {
	return m.withSession(session).SoftDeleteByIDs(ctx, ids, deletedAt, modifier)
}
