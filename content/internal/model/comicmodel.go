package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ComicModel = (*customComicModel)(nil)

const (
	comicTagTable = "comic_tag"
)

type (
	// ComicModel is an interface to be customized, add more methods here,
	// and implement the added methods in customComicModel.
	ComicModel interface {
		comicModel
		withSession(session sqlx.Session) ComicModel
		IncrChapterCountByComicID(ctx context.Context, comicID uint64) error
		IncrChapterCountByComicIDWithSession(ctx context.Context, session sqlx.Session, comicID uint64) error
		IncrCommentCount(ctx context.Context, comicID uint64) (sql.Result, error)
		IncrCommentCountWithSession(ctx context.Context, session sqlx.Session, comicID uint64) (sql.Result, error)
		DecrCommentCount(ctx context.Context, comicID uint64) (sql.Result, error)
		DecrCommentCountWithSession(ctx context.Context, session sqlx.Session, comicID uint64) (sql.Result, error)
		DecrCommentCountByCount(ctx context.Context, comicID uint64, count uint64) (sql.Result, error)
		DecrCommentCountByCountWithSession(ctx context.Context, session sqlx.Session, comicID uint64, count uint64) (sql.Result, error)
		DecrChapterCountByComicID(ctx context.Context, comicID uint64) error
		DecrChapterCountByComicIDWithSession(ctx context.Context, session sqlx.Session, comicID uint64) error
		IncrLikeCount(ctx context.Context, id uint64) (sql.Result, error)
		IncrLikeCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
		DecrLikeCount(ctx context.Context, id uint64) (sql.Result, error)
		DecrLikeCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
		IncrViewCount(ctx context.Context, id uint64) (sql.Result, error)
		IncrViewCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
		FindOneWithNotDelete(ctx context.Context, id uint64) (*Comic, error)
		FindOneWithNotDeleteWithSession(ctx context.Context, session sqlx.Session, id uint64) (*Comic, error)
		FindByTagsAndKeyWord(ctx context.Context, offset int, limit int, tags []string, keyword string) ([]*Comic, error)
		UpdateWithSession(ctx context.Context, session sqlx.Session, data *Comic) error
		UpdateRelationStatus(ctx context.Context, id uint64, relationStatus int64) error
		UpdateRelationStatusWithSession(ctx context.Context, session sqlx.Session, id uint64, relationStatus int64) error
		SoftDelete(ctx context.Context, id uint64, deletedAt uint64, modifier sql.NullInt64) error
	}

	customComicModel struct {
		*defaultComicModel
	}
)

// NewComicModel returns a model for the database table.
func NewComicModel(conn sqlx.SqlConn) ComicModel {
	return &customComicModel{
		defaultComicModel: newComicModel(conn),
	}
}

func (m *customComicModel) withSession(session sqlx.Session) ComicModel {
	return NewComicModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customComicModel) IncrChapterCountByComicID(ctx context.Context, comicID uint64) error {
	query := fmt.Sprintf("update %s set `chapter_count` = `chapter_count` + 1 where `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, comicID)
	return mapDBError(err)
}

func (m *customComicModel) IncrChapterCountByComicIDWithSession(ctx context.Context, session sqlx.Session, comicID uint64) error {
	return m.withSession(session).IncrChapterCountByComicID(ctx, comicID)
}

func (m *customComicModel) IncrCommentCount(ctx context.Context, comicID uint64) (sql.Result, error) {
	query := fmt.Sprintf("update %s set `comment_count` = `comment_count` + 1 where `id` = ? and `deleted_at` = 0", m.table)
	result, err := m.conn.ExecCtx(ctx, query, comicID)
	return result, mapDBError(err)
}

func (m *customComicModel) IncrCommentCountWithSession(ctx context.Context, session sqlx.Session, comicID uint64) (sql.Result, error) {
	return m.withSession(session).IncrCommentCount(ctx, comicID)
}

func (m *customComicModel) DecrCommentCount(ctx context.Context, comicID uint64) (sql.Result, error) {
	query := fmt.Sprintf("update %s set `comment_count` = `comment_count` - 1 where `id` = ? and `comment_count` > 0 and `deleted_at` = 0", m.table)
	result, err := m.conn.ExecCtx(ctx, query, comicID)
	return result, mapDBError(err)
}

func (m *customComicModel) DecrCommentCountWithSession(ctx context.Context, session sqlx.Session, comicID uint64) (sql.Result, error) {
	return m.withSession(session).DecrCommentCount(ctx, comicID)
}

func (m *customComicModel) DecrCommentCountByCount(ctx context.Context, comicID uint64, count uint64) (sql.Result, error) {
	query := fmt.Sprintf("update %s set `comment_count` = `comment_count` - ? where `id` = ? and `comment_count` >= ? and `deleted_at` = 0", m.table)
	result, err := m.conn.ExecCtx(ctx, query, count, comicID, count)
	return result, mapDBError(err)
}

func (m *customComicModel) DecrCommentCountByCountWithSession(ctx context.Context, session sqlx.Session, comicID uint64, count uint64) (sql.Result, error) {
	return m.withSession(session).DecrCommentCountByCount(ctx, comicID, count)
}

func (m *customComicModel) DecrChapterCountByComicID(ctx context.Context, comicID uint64) error {
	query := fmt.Sprintf("update %s set `chapter_count` = `chapter_count` - 1 where `id` = ? and `chapter_count` > 0", m.table)
	_, err := m.conn.ExecCtx(ctx, query, comicID)
	return mapDBError(err)
}

func (m *customComicModel) DecrChapterCountByComicIDWithSession(ctx context.Context, session sqlx.Session, comicID uint64) error {
	return m.withSession(session).DecrChapterCountByComicID(ctx, comicID)
}

func (m *customComicModel) IncrLikeCount(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s 
		set like_count = like_count + 1 
		where id = ? and deleted_at = 0`,
		m.table)

	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customComicModel) IncrLikeCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).IncrLikeCount(ctx, id)
}

func (m *customComicModel) DecrLikeCount(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set like_count = like_count - 1
		where id = ? and like_count > 0 and deleted_at = 0`,
		m.table,
	)

	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customComicModel) DecrLikeCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).DecrLikeCount(ctx, id)
}

func (m *customComicModel) IncrViewCount(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set view_count = view_count + 1
		where id = ? and deleted_at = 0`,
		m.table,
	)

	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customComicModel) IncrViewCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).IncrViewCount(ctx, id)
}

func (m *customComicModel) FindOneWithNotDelete(ctx context.Context, id uint64) (*Comic, error) {
	query := fmt.Sprintf(
		"select %s from %s where `id` = ? and `deleted_at` = 0 limit 1",
		comicRows,
		m.table,
	)

	var resp Comic
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	return &resp, mapDBError(err)
}

func (m *customComicModel) FindOneWithNotDeleteWithSession(ctx context.Context, session sqlx.Session, id uint64) (*Comic, error) {
	return m.withSession(session).FindOneWithNotDelete(ctx, id)
}

func (m *customComicModel) FindByTagsAndKeyWord(ctx context.Context, offset int, limit int, tags []string, keyword string) ([]*Comic, error) {
	kw := "%" + keyword + "%"
	args := make([]interface{}, 0, len(tags)+5) // 占位符的数据，后两个是 offest 和 limit
	args = append(args, kw, kw, kw)             // name, description, author
	tagFilter := ""                             // tags 为空时跳过筛选
	if len(tags) > 0 {
		placeholders := make([]string, 0, len(tags))
		for _, tag := range tags {
			placeholders = append(placeholders, "?")
			args = append(args, tag)
		}
		tagFilter = "and t.tag in (" + strings.Join(placeholders, ",") + ")"
	}
	args = append(args, offset, limit)

	// distinct 去重
	query := fmt.Sprintf(`
		select distinct c.*
		from %s c
		left join %s t on c.id = t.comic_id
		where (c.name like ? or c.description like ? or c.author like ?)
		%s
		limit ?,?
`, m.table, comicTagTable, tagFilter)

	var out []*Comic
	err := m.conn.QueryRowsCtx(ctx, &out, query, args...)
	return out, mapDBError(err)
}

func (m *customComicModel) UpdateWithSession(ctx context.Context, session sqlx.Session, data *Comic) error {
	return m.withSession(session).Update(ctx, data)
}

func (m *customComicModel) UpdateRelationStatus(ctx context.Context, id uint64, relationStatus int64) error {
	query := fmt.Sprintf("update %s set `relation_status` = ? where `id` = ?", m.table)

	_, err := m.conn.ExecCtx(ctx, query, relationStatus, id)
	return mapDBError(err)
}

func (m *customComicModel) UpdateRelationStatusWithSession(ctx context.Context, session sqlx.Session, id uint64, relationStatus int64) error {
	return m.withSession(session).UpdateRelationStatus(ctx, id, relationStatus)
}

func (m *customComicModel) SoftDelete(ctx context.Context, id uint64, deletedAt uint64, modifier sql.NullInt64) error {
	query := fmt.Sprintf(
		"update %s set `deleted_at` = ?, `last_modified_by` = ? where `id` = ? and `deleted_at` = 0",
		m.table,
	)
	_, err := m.conn.ExecCtx(ctx, query, deletedAt, modifier, id)
	return mapDBError(err)
}
