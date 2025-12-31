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
		InsertBatchWithSession(ctx context.Context, session sqlx.Session, list []*ArticleTag) error
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
