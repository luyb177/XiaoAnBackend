package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ArticleModel = (*customArticleModel)(nil)

const (
	articleTagTable = "article_tag"
)

type (
	// ArticleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customArticleModel.
	ArticleModel interface {
		articleModel
		withSession(session sqlx.Session) ArticleModel
		InsertWithSession(ctx context.Context, session sqlx.Session, data *Article) (sql.Result, error)
		IncrCommentCount(ctx context.Context, id uint64) (sql.Result, error)
		IncrCommentCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
		DecrCommentCount(ctx context.Context, id uint64) (sql.Result, error)
		DecrCommentCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
		DecrCommentCountByCount(ctx context.Context, id uint64, count uint64) (sql.Result, error)
		DecrCommentCountByCountWithSession(ctx context.Context, session sqlx.Session, id uint64, count uint64) (sql.Result, error)
		IncrLikeCount(ctx context.Context, id uint64) (sql.Result, error)
		IncrLikeCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
		DecrLikeCount(ctx context.Context, id uint64) (sql.Result, error)
		DecrLikeCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
		IncrViewCount(ctx context.Context, id uint64) (sql.Result, error)
		IncrViewCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
		FindByTagsAndKeyWord(ctx context.Context, offset int, limit int, tags []string, keyword string) ([]*Article, error)
		FindOneWithNotDelete(ctx context.Context, id uint64) (*Article, error)
		FindOneWithNotDeleteWithSession(ctx context.Context, session sqlx.Session, id uint64) (*Article, error)
		UpdateWithSession(ctx context.Context, session sqlx.Session, data *Article) error
		UpdateRelationStatus(ctx context.Context, id uint64, relationStatus int64) error
		UpdateRelationStatusWithSession(ctx context.Context, session sqlx.Session, id uint64, relationStatus int64) error
		SoftDelete(ctx context.Context, id uint64, deletedAt uint64, modifier sql.NullInt64) error
	}

	customArticleModel struct {
		*defaultArticleModel
	}
)

// NewArticleModel returns a model for the database table.
func NewArticleModel(conn sqlx.SqlConn) ArticleModel {
	return &customArticleModel{
		defaultArticleModel: newArticleModel(conn),
	}
}

func (m *customArticleModel) withSession(session sqlx.Session) ArticleModel {
	return NewArticleModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customArticleModel) InsertWithSession(ctx context.Context, session sqlx.Session, data *Article) (sql.Result, error) {
	return m.withSession(session).Insert(ctx, data)
}

func (m *customArticleModel) IncrCommentCount(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(
		"update %s set `comment_count` = `comment_count` + 1 where `id` = ? and deleted_at = 0",
		m.table,
	)
	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customArticleModel) IncrCommentCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).IncrCommentCount(ctx, id)
}

func (m *customArticleModel) DecrCommentCount(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(
		"update %s set `comment_count` = `comment_count` - 1 where id = ? and comment_count > 0 and deleted_at = 0",
		m.table,
	)
	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customArticleModel) DecrCommentCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).DecrCommentCount(ctx, id)
}

func (m *customArticleModel) DecrCommentCountByCount(ctx context.Context, id uint64, count uint64) (sql.Result, error) {
	query := fmt.Sprintf(
		"update %s set `comment_count` = `comment_count` - ? where id = ? and comment_count >= ? and deleted_at = 0",
		m.table,
	)
	result, err := m.conn.ExecCtx(ctx, query, count, id, count)
	return result, mapDBError(err)
}

func (m *customArticleModel) DecrCommentCountByCountWithSession(ctx context.Context, session sqlx.Session, id uint64, count uint64) (sql.Result, error) {
	return m.withSession(session).DecrCommentCountByCount(ctx, id, count)
}

func (m *customArticleModel) IncrLikeCount(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(
		"update %s set `like_count` = `like_count` + 1 where `id` = ? and deleted_at = 0",
		m.table,
	)
	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customArticleModel) IncrLikeCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).IncrLikeCount(ctx, id)
}

func (m *customArticleModel) DecrLikeCount(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set like_count = like_count - 1 
		where id = ? and like_count > 0 and deleted_at = 0`,
		m.table)

	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customArticleModel) DecrLikeCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).DecrLikeCount(ctx, id)
}

func (m *customArticleModel) IncrViewCount(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set view_count = view_count + 1
		where id = ? and deleted_at = 0`,
		m.table,
	)

	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customArticleModel) IncrViewCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).IncrViewCount(ctx, id)
}

func (m *customArticleModel) FindByTagsAndKeyWord(ctx context.Context, offset int, limit int, tags []string, keyword string) ([]*Article, error) {
	kw := "%" + keyword + "%"

	args := make([]interface{}, 0, len(tags)+6)
	args = append(args, kw, kw, kw, kw) // name, description, author,content

	tagFilter := ""
	if len(tags) > 0 {
		placeholders := make([]string, 0, len(tags))
		for _, tag := range tags {
			placeholders = append(placeholders, "?")
			args = append(args, tag)
		}
		tagFilter = "and t.tag in" + "(" + strings.Join(placeholders, ",") + ")"
	}

	args = append(args, offset, limit)

	query := fmt.Sprintf(`
		select distinct a.*
		from %s a 
		left join %s t on a.id = t.article_id
		where (a.name like ? or a.description like ? or a.author like ? or a.content like ?)
		%s
		limit ?,?`, m.table, articleTagTable, tagFilter)

	var out []*Article
	err := m.conn.QueryRowsCtx(ctx, &out, query, args...)
	return out, mapDBError(err)
}

func (m *customArticleModel) FindOneWithNotDelete(ctx context.Context, id uint64) (*Article, error) {
	query := fmt.Sprintf(
		"select %s from %s where `id` = ? and `deleted_at` = 0 limit 1",
		articleRows,
		m.table,
	)

	var resp Article
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	return &resp, mapDBError(err)
}

func (m *customArticleModel) FindOneWithNotDeleteWithSession(ctx context.Context, session sqlx.Session, id uint64) (*Article, error) {
	return m.withSession(session).FindOneWithNotDelete(ctx, id)
}

func (m *customArticleModel) UpdateWithSession(ctx context.Context, session sqlx.Session, data *Article) error {
	return m.withSession(session).Update(ctx, data)
}

func (m *customArticleModel) UpdateRelationStatus(ctx context.Context, id uint64, relationStatus int64) error {
	query := fmt.Sprintf("update %s set `relation_status` = ? where `id` = ?", m.table)

	_, err := m.conn.ExecCtx(ctx, query, relationStatus, id)
	return mapDBError(err)
}

func (m *customArticleModel) UpdateRelationStatusWithSession(ctx context.Context, session sqlx.Session, id uint64, relationStatus int64) error {
	return m.withSession(session).UpdateRelationStatus(ctx, id, relationStatus)
}

func (m *customArticleModel) SoftDelete(ctx context.Context, id uint64, deletedAt uint64, modifier sql.NullInt64) error {
	query := fmt.Sprintf(
		"update %s set `deleted_at` = ?, `last_modified_by` = ? where `id` = ?",
		m.table,
	)

	_, err := m.conn.ExecCtx(ctx, query, deletedAt, modifier, id)
	return mapDBError(err)
}
