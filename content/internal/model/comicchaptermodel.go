package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ComicChapterModel = (*customComicChapterModel)(nil)

type (
	// ComicChapterModel is an interface to be customized, add more methods here,
	// and implement the added methods in customComicChapterModel.
	ComicChapterModel interface {
		comicChapterModel
		withSession(session sqlx.Session) ComicChapterModel
		CustomInsert(ctx context.Context, data *ComicChapter) (sql.Result, error)
		CustomInsertWithSession(ctx context.Context, session sqlx.Session, data *ComicChapter) (sql.Result, error)
		FindOneWithNotDelete(ctx context.Context, id uint64) (*ComicChapter, error)
		FindOneByComicIDAndChapterNo(ctx context.Context, comicID uint64, chapterNo int64) (*ComicChapter, error)
		FindManyByComicIDOrderByChapterNo(ctx context.Context, comicID uint64, offset, pageSize int64) ([]*ComicChapter, error)
		FindAllByComicID(ctx context.Context, comicID uint64) ([]*ComicChapter, error)
		UpdateRelationStatus(ctx context.Context, id uint64, relationStatus int64) error
		UpdateRelationStatusWithSession(ctx context.Context, session sqlx.Session, id uint64, relationStatus int64) error
		SoftDelete(ctx context.Context, id uint64, uid int64) error
		SoftDeleteByIDs(ctx context.Context, ids []uint64, uid int64) error
		SoftDeleteByIDsWithSession(ctx context.Context, session sqlx.Session, ids []uint64, uid int64) error
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

func (m *customComicChapterModel) CustomInsert(ctx context.Context, data *ComicChapter) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, comicChapterRowsExpectAutoSet)
	ret, err := m.conn.ExecCtx(ctx, query, data.ComicId, data.ChapterNo, data.Title, data.Description, data.PageCount, data.Status, data.PublishedAt, data.RelationStatus, data.LastModifiedBy, data.DeletedAt)
	return ret, mapDBError(err)
}

func (m *customComicChapterModel) CustomInsertWithSession(ctx context.Context, session sqlx.Session, data *ComicChapter) (sql.Result, error) {
	return m.withSession(session).CustomInsert(ctx, data)
}

func (m *customComicChapterModel) FindOneWithNotDelete(ctx context.Context, id uint64) (*ComicChapter, error) {
	query := fmt.Sprintf(
		"select %s from %s where `id` = ? and `deleted_at` = 0 limit 1",
		comicChapterRows,
		m.table,
	)

	var resp ComicChapter
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlc.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customComicChapterModel) FindOneByComicIDAndChapterNo(ctx context.Context, comicID uint64, chapterNo int64) (*ComicChapter, error) {
	query := fmt.Sprintf(
		"select %s from %s where `comic_id` = ? and `chapter_no` = ? and `deleted_at` = 0 limit 1",
		comicChapterRows,
		m.table,
	)

	var resp ComicChapter
	err := m.conn.QueryRowCtx(ctx, &resp, query, comicID, chapterNo)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlc.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customComicChapterModel) FindManyByComicIDOrderByChapterNo(ctx context.Context, comicID uint64, offset, pageSize int64) ([]*ComicChapter, error) {
	query := fmt.Sprintf(
		"select %s from %s where `comic_id` = ? and `deleted_at` = 0 order by `chapter_no` asc limit ? offset ?",
		comicChapterRows,
		m.table,
	)

	var resp []*ComicChapter
	err := m.conn.QueryRowsCtx(ctx, &resp, query, comicID, pageSize, offset)
	return resp, err
}

func (m *customComicChapterModel) FindAllByComicID(ctx context.Context, comicID uint64) ([]*ComicChapter, error) {
	query := fmt.Sprintf(
		"select %s from %s where `comic_id` = ? and `deleted_at` = 0",
		comicChapterRows,
		m.table,
	)

	var resp []*ComicChapter
	err := m.conn.QueryRowsCtx(ctx, &resp, query, comicID)
	return resp, err
}

func (m *customComicChapterModel) UpdateRelationStatus(ctx context.Context, id uint64, relationStatus int64) error {
	query := fmt.Sprintf("update %s set `relation_status` = ? where `id` = ?", m.table)

	_, err := m.conn.ExecCtx(ctx, query, relationStatus, id)
	return err
}

func (m *customComicChapterModel) UpdateRelationStatusWithSession(ctx context.Context, session sqlx.Session, id uint64, relationStatus int64) error {
	return m.withSession(session).UpdateRelationStatus(ctx, id, relationStatus)
}

func (m *customComicChapterModel) SoftDelete(ctx context.Context, id uint64, uid int64) error {
	deletedAt := uint64(time.Now().Unix())
	modifier := sql.NullInt64{Int64: uid, Valid: true}

	query := fmt.Sprintf(`
		UPDATE %s
		SET deleted_at = ?, last_modified_by = ?
		WHERE id = ? AND deleted_at = 0
	`, m.table)

	_, err := m.conn.ExecCtx(ctx, query, deletedAt, modifier, id)
	return err
}

func (m *customComicChapterModel) SoftDeleteByIDs(ctx context.Context, ids []uint64, uid int64) error {
	if len(ids) == 0 {
		return nil
	}

	deletedAt := uint64(time.Now().Unix())
	modifier := sql.NullInt64{Int64: uid, Valid: true}

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
	return err
}

func (m *customComicChapterModel) SoftDeleteByIDsWithSession(ctx context.Context, session sqlx.Session, ids []uint64, uid int64) error {
	return m.withSession(session).SoftDeleteByIDs(ctx, ids, uid)
}
