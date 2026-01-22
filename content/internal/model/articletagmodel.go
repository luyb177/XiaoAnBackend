package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ArticleTagModel = (*customArticleTagModel)(nil)

type (
	// ArticleTagModel is an interface to be customized, add more methods here,
	// and implement the added methods in customArticleTagModel.
	ArticleTagModel interface {
		articleTagModel
		withSession(session sqlx.Session) ArticleTagModel
		InsertBatch(ctx context.Context, list []*ArticleTag) error
		InsertBatchWithSession(ctx context.Context, session sqlx.Session, list []*ArticleTag) error
		FindManyByArticleId(ctx context.Context, articleId uint64) ([]*ArticleTag, error)
		DeleteBatchByArticleId(ctx context.Context, articleId uint64) error
		DeleteBatchByArticleIdWithSession(ctx context.Context, session sqlx.Session, articleId uint64) error
	}

	customArticleTagModel struct {
		*defaultArticleTagModel
	}
)

// NewArticleTagModel returns a model for the database table.
func NewArticleTagModel(conn sqlx.SqlConn) ArticleTagModel {
	return &customArticleTagModel{
		defaultArticleTagModel: newArticleTagModel(conn),
	}
}

func (m *customArticleTagModel) withSession(session sqlx.Session) ArticleTagModel {
	return NewArticleTagModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customArticleTagModel) InsertBatch(ctx context.Context, list []*ArticleTag) error {
	if len(list) == 0 {
		return nil
	}

	// 构造 values
	valuePlaceholders := make([]string, 0, len(list))
	args := make([]interface{}, 0, len(list)*3)

	for _, tag := range list {
		valuePlaceholders = append(valuePlaceholders, "(?,?,?)")
		args = append(args, tag.ArticleId, tag.Tag, tag.DeletedAt)
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES %s`,
		m.table,
		articleTagRowsExpectAutoSet,
		strings.Join(valuePlaceholders, ","),
	)

	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}

func (m *customArticleTagModel) InsertBatchWithSession(ctx context.Context, session sqlx.Session, list []*ArticleTag) error {
	if len(list) == 0 {
		return nil
	}

	// 构造 values
	valuePlaceholders := make([]string, 0, len(list))
	args := make([]interface{}, 0, len(list)*3)

	for _, tag := range list {
		valuePlaceholders = append(valuePlaceholders, "(?,?,?)")
		args = append(args, tag.ArticleId, tag.Tag, tag.DeletedAt)
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES %s`,
		m.table,
		articleTagRowsExpectAutoSet,
		strings.Join(valuePlaceholders, ","),
	)

	_, err := session.ExecCtx(ctx, query, args...)
	return err
}

func (m *customArticleTagModel) FindManyByArticleId(ctx context.Context, articleId uint64) ([]*ArticleTag, error) {
	query := fmt.Sprintf(
		"select %s from %s where `article_id` = ?",
		articleTagRows,
		m.table,
	)

	var resp []*ArticleTag
	err := m.conn.QueryRowsCtx(ctx, &resp, query, articleId)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (m *customArticleTagModel) DeleteBatchByArticleIdWithSession(ctx context.Context, session sqlx.Session, articleId uint64) error {
	return m.withSession(session).DeleteBatchByArticleId(ctx, articleId)
}

func (m *customArticleTagModel) DeleteBatchByArticleId(ctx context.Context, articleId uint64) error {
	query := fmt.Sprintf(
		"delete from %s where `article_id` = ?",
		m.table,
	)
	_, err := m.conn.ExecCtx(ctx, query, articleId)
	return err
}
