package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ComicUrlModel = (*customComicUrlModel)(nil)

type (
	// ComicUrlModel is an interface to be customized, add more methods here,
	// and implement the added methods in customComicUrlModel.
	ComicUrlModel interface {
		comicUrlModel
		withSession(session sqlx.Session) ComicUrlModel
	}

	customComicUrlModel struct {
		*defaultComicUrlModel
	}
)

// NewComicUrlModel returns a model for the database table.
func NewComicUrlModel(conn sqlx.SqlConn) ComicUrlModel {
	return &customComicUrlModel{
		defaultComicUrlModel: newComicUrlModel(conn),
	}
}

func (m *customComicUrlModel) withSession(session sqlx.Session) ComicUrlModel {
	return NewComicUrlModel(sqlx.NewSqlConnFromSession(session))
}
