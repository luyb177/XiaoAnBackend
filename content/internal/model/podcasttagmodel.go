package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ PodcastTagModel = (*customPodcastTagModel)(nil)

type (
	// PodcastTagModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPodcastTagModel.
	PodcastTagModel interface {
		podcastTagModel
		withSession(session sqlx.Session) PodcastTagModel
		InsertBatch(ctx context.Context, list []*PodcastTag) error
		InsertBatchWithSession(ctx context.Context, session sqlx.Session, list []*PodcastTag) error
		FindManyByPodcastId(ctx context.Context, podcastId uint64) ([]*PodcastTag, error)
		DeleteBatchByPodcastId(ctx context.Context, podcastId uint64) error
		DeleteBatchByPodcastIdWithSession(ctx context.Context, session sqlx.Session, podcastId uint64) error
		SoftDeleteByPodcastId(ctx context.Context, podcastId uint64, deletedAt uint64) error
		SoftDeleteByPodcastIdWithSession(ctx context.Context, session sqlx.Session, podcastId uint64, deletedAt uint64) error
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

func (m *customPodcastTagModel) InsertBatch(ctx context.Context, list []*PodcastTag) error {
	if len(list) == 0 {
		return nil
	}

	// 构造 values
	valuePlaceholders := make([]string, 0, len(list))
	args := make([]interface{}, 0, len(list)*3)

	for _, tag := range list {
		valuePlaceholders = append(valuePlaceholders, "(?,?,?)")
		args = append(args, tag.PodcastId, tag.Tag, tag.DeletedAt)
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES %s`,
		m.table,
		podcastTagRowsExpectAutoSet,
		strings.Join(valuePlaceholders, ","),
	)

	_, err := m.conn.ExecCtx(ctx, query, args...)
	return mapDBError(err)
}

func (m *customPodcastTagModel) InsertBatchWithSession(ctx context.Context, session sqlx.Session, list []*PodcastTag) error {
	return m.withSession(session).InsertBatch(ctx, list)
}

func (m *customPodcastTagModel) FindManyByPodcastId(ctx context.Context, podcastId uint64) ([]*PodcastTag, error) {
	query := fmt.Sprintf(
		"select %s from %s where `podcast_id` = ?",
		podcastTagRows,
		m.table,
	)

	var resp []*PodcastTag
	err := m.conn.QueryRowsCtx(ctx, &resp, query, podcastId)
	return resp, mapDBError(err)
}

func (m *customPodcastTagModel) DeleteBatchByPodcastId(ctx context.Context, podcastId uint64) error {
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE `podcast_id` = ?",
		m.table,
	)

	_, err := m.conn.ExecCtx(ctx, query, podcastId)
	return mapDBError(err)
}

func (m *customPodcastTagModel) DeleteBatchByPodcastIdWithSession(ctx context.Context, session sqlx.Session, podcastId uint64) error {
	return m.withSession(session).DeleteBatchByPodcastId(ctx, podcastId)
}

func (m *customPodcastTagModel) SoftDeleteByPodcastId(ctx context.Context, podcastId uint64, deletedAt uint64) error {
	query := fmt.Sprintf(
		"UPDATE %s SET `deleted_at` = ? WHERE `podcast_id` = ?",
		m.table,
	)

	_, err := m.conn.ExecCtx(ctx, query, deletedAt, podcastId)
	return mapDBError(err)
}

func (m *customPodcastTagModel) SoftDeleteByPodcastIdWithSession(ctx context.Context, session sqlx.Session, podcastId uint64, deletedAt uint64) error {
	return m.withSession(session).SoftDeleteByPodcastId(ctx, podcastId, deletedAt)
}
