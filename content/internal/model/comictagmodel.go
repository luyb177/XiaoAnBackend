package model

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ComicTagModel = (*customComicTagModel)(nil)

type (
	// ComicTagModel is an interface to be customized, add more methods here,
	// and implement the added methods in customComicTagModel.
	ComicTagModel interface {
		comicTagModel
		withSession(session sqlx.Session) ComicTagModel
	}

	customComicTagModel struct {
		*defaultComicTagModel
	}
)

// NewComicTagModel returns a model for the database table.
func NewComicTagModel(conn sqlx.SqlConn) ComicTagModel {
	return &customComicTagModel{
		defaultComicTagModel: newComicTagModel(conn),
	}
}

func (m *customComicTagModel) withSession(session sqlx.Session) ComicTagModel {
	return NewComicTagModel(sqlx.NewSqlConnFromSession(session))
}
