package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ PodcastHighlightModel = (*customPodcastHighlightModel)(nil)

type (
	// PodcastHighlightModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPodcastHighlightModel.
	PodcastHighlightModel interface {
		podcastHighlightModel
		withSession(session sqlx.Session) PodcastHighlightModel
		InsertBatch(ctx context.Context, list []*PodcastHighlight) error
		InsertBatchWithSession(ctx context.Context, session sqlx.Session, list []*PodcastHighlight) error
		FindManyByPodcastId(ctx context.Context, podcastId uint64) ([]*PodcastHighlight, error)
		DeleteBatchByPodcastId(ctx context.Context, podcastId uint64) error
		DeleteBatchByPodcastIdWithSession(ctx context.Context, session sqlx.Session, podcastId uint64) error
	}

	customPodcastHighlightModel struct {
		*defaultPodcastHighlightModel
	}
)

// NewPodcastHighlightModel returns a model for the database table.
func NewPodcastHighlightModel(conn sqlx.SqlConn) PodcastHighlightModel {
	return &customPodcastHighlightModel{
		defaultPodcastHighlightModel: newPodcastHighlightModel(conn),
	}
}

func (m *customPodcastHighlightModel) withSession(session sqlx.Session) PodcastHighlightModel {
	return NewPodcastHighlightModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customPodcastHighlightModel) InsertBatch(ctx context.Context, list []*PodcastHighlight) error {
	if len(list) == 0 {
		return nil
	}

	// 构造 values
	valuePlaceholders := make([]string, 0, len(list))
	args := make([]interface{}, 0, len(list)*4)

	for _, highlight := range list {
		valuePlaceholders = append(valuePlaceholders, "(?,?,?,?)")
		args = append(args, highlight.PodcastId, highlight.Second, highlight.Highlight, highlight.DeletedAt)
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES %s`,
		m.table,
		podcastHighlightRowsExpectAutoSet,
		strings.Join(valuePlaceholders, ","),
	)

	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}
func (m *customPodcastHighlightModel) InsertBatchWithSession(ctx context.Context, session sqlx.Session, list []*PodcastHighlight) error {
	return m.withSession(session).InsertBatch(ctx, list)
}

func (m *customPodcastHighlightModel) FindManyByPodcastId(ctx context.Context, podcastId uint64) ([]*PodcastHighlight, error) {
	query := fmt.Sprintf(
		"select %s from %s where `podcast_id` = ?",
		podcastHighlightRows,
		m.table,
	)
	var resp []*PodcastHighlight
	err := m.conn.QueryRowsCtx(ctx, &resp, query, podcastId)
	return resp, err
}

func (m *customPodcastHighlightModel) DeleteBatchByPodcastId(ctx context.Context, podcastId uint64) error {
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE `podcast_id` = ?",
		m.table,
	)

	_, err := m.conn.ExecCtx(ctx, query, podcastId)
	return err
}

func (m *customPodcastHighlightModel) DeleteBatchByPodcastIdWithSession(ctx context.Context, session sqlx.Session, podcastId uint64) error {
	return m.withSession(session).DeleteBatchByPodcastId(ctx, podcastId)
}
