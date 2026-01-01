package model

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
)

var _ ArticleImageModel = (*customArticleImageModel)(nil)

type (
	// ArticleImageModel is an interface to be customized, add more methods here,
	// and implement the added methods in customArticleImageModel.
	ArticleImageModel interface {
		articleImageModel
		withSession(session sqlx.Session) ArticleImageModel
		InsertBatchWithSession(ctx context.Context, session sqlx.Session, list []*ArticleImage) error
		FindManyByArticleId(ctx context.Context, articleId uint64) ([]*ArticleImage, error)
	}

	customArticleImageModel struct {
		*defaultArticleImageModel
	}
)

// NewArticleImageModel returns a model for the database table.
func NewArticleImageModel(conn sqlx.SqlConn) ArticleImageModel {
	return &customArticleImageModel{
		defaultArticleImageModel: newArticleImageModel(conn),
	}
}

func (m *customArticleImageModel) withSession(session sqlx.Session) ArticleImageModel {
	return NewArticleImageModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customArticleImageModel) InsertBatchWithSession(ctx context.Context, session sqlx.Session, list []*ArticleImage) error {
	if len(list) == 0 {
		return nil
	}

	// 构造 values (?, ?, ?, ?, ?), (?, ?, ?, ?, ?)
	valuePlaceholders := make([]string, 0, len(list))
	args := make([]interface{}, 0, len(list)*5)

	for _, img := range list {
		valuePlaceholders = append(valuePlaceholders, "(?, ?, ?, ?, ?)")
		args = append(args,
			img.ArticleId,
			img.Url,
			img.Sort,
			img.Type,
			img.DeletedAt,
		)
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES %s`,
		m.table,
		articleImageRowsExpectAutoSet,
		strings.Join(valuePlaceholders, ","),
	)

	_, err := session.ExecCtx(ctx, query, args...)
	return err
}

func (m *customArticleImageModel) FindManyByArticleId(ctx context.Context, articleId uint64) ([]*ArticleImage, error) {
	query := fmt.Sprintf("select %s from %s where `article_id` = ?", articleImageRows, m.table)

	var res []*ArticleImage

	err := m.conn.QueryRowsCtx(ctx, &res, query, articleId)
	if err != nil {
		return nil, err
	}
	return res, nil
}
