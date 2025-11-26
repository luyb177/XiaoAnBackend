package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ContentLikeModel = (*customContentLikeModel)(nil)

type (
	// ContentLikeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customContentLikeModel.
	ContentLikeModel interface {
		contentLikeModel
		withSession(session sqlx.Session) ContentLikeModel
	}

	customContentLikeModel struct {
		*defaultContentLikeModel
	}
)

// NewContentLikeModel returns a model for the database table.
func NewContentLikeModel(conn sqlx.SqlConn) ContentLikeModel {
	return &customContentLikeModel{
		defaultContentLikeModel: newContentLikeModel(conn),
	}
}

func (m *customContentLikeModel) withSession(session sqlx.Session) ContentLikeModel {
	return NewContentLikeModel(sqlx.NewSqlConnFromSession(session))
}
