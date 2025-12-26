package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ArticleTagModel = (*customArticleTagModel)(nil)

type (
	// ArticleTagModel is an interface to be customized, add more methods here,
	// and implement the added methods in customArticleTagModel.
	ArticleTagModel interface {
		articleTagModel
		withSession(session sqlx.Session) ArticleTagModel
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
