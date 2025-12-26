package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ PodcastTagModel = (*customPodcastTagModel)(nil)

type (
	// PodcastTagModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPodcastTagModel.
	PodcastTagModel interface {
		podcastTagModel
		withSession(session sqlx.Session) PodcastTagModel
	}

	customPodcastTagModel struct {
		*defaultPodcastTagModel
	}
)

// NewPodcastTagModel returns a model for the database table.
func NewPodcastTagModel(conn sqlx.SqlConn) PodcastTagModel {
	return &customPodcastTagModel{
		defaultPodcastTagModel: newPodcastTagModel(conn),
	}
}

func (m *customPodcastTagModel) withSession(session sqlx.Session) PodcastTagModel {
	return NewPodcastTagModel(sqlx.NewSqlConnFromSession(session))
}
