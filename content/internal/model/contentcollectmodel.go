package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ContentCollectModel = (*customContentCollectModel)(nil)

type (
	// ContentCollectModel is an interface to be customized, add more methods here,
	// and implement the added methods in customContentCollectModel.
	ContentCollectModel interface {
		contentCollectModel
		withSession(session sqlx.Session) ContentCollectModel
	}

	customContentCollectModel struct {
		*defaultContentCollectModel
	}
)

// NewContentCollectModel returns a model for the database table.
func NewContentCollectModel(conn sqlx.SqlConn) ContentCollectModel {
	return &customContentCollectModel{
		defaultContentCollectModel: newContentCollectModel(conn),
	}
}

func (m *customContentCollectModel) withSession(session sqlx.Session) ContentCollectModel {
	return NewContentCollectModel(sqlx.NewSqlConnFromSession(session))
}
