package model

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
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
		FindByTagsAndKeyWord(ctx context.Context, offset int, limit int, tags []string, keyword string) ([]*Article, error)
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
		seletct distinct a.*
		from %s a 
		left join %s t on a.id = t.article_id
		where (a.name like ? or a.description like ? or a.author like ? or a.content like ?)
		%s
		limit ?,?`, m.table, articleTagTable, tagFilter)

	var out []*Article
	err := m.conn.QueryRowsCtx(ctx, &out, query, args...)
	return out, err
}
