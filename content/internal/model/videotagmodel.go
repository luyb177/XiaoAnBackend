package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ VideoTagModel = (*customVideoTagModel)(nil)

type (
	// VideoTagModel is an interface to be customized, add more methods here,
	// and implement the added methods in customVideoTagModel.
	VideoTagModel interface {
		videoTagModel
		withSession(session sqlx.Session) VideoTagModel
		InsertBatch(ctx context.Context, list []*VideoTag) error
		InsertBatchWithSession(ctx context.Context, session sqlx.Session, list []*VideoTag) error
		FindManyByVideoID(ctx context.Context, videoID uint64) ([]*VideoTag, error)
		FindByVideoTags(ctx context.Context, offest, limit int, tags []string) ([]*VideoTag, error)
		DeleteBatchByVideoID(ctx context.Context, videoID uint64) error
		DeleteBatchByVideoIDWithSession(ctx context.Context, session sqlx.Session, videoID uint64) error
		SoftDeleteByVideoID(ctx context.Context, videoID, deletedAt uint64) error
		SoftDeleteByVideoIDWithSession(ctx context.Context, session sqlx.Session, videoID, deletedAt uint64) error
	}

	customVideoTagModel struct {
		*defaultVideoTagModel
	}
)

// NewVideoTagModel returns a model for the database table.
func NewVideoTagModel(conn sqlx.SqlConn) VideoTagModel {
	return &customVideoTagModel{
		defaultVideoTagModel: newVideoTagModel(conn),
	}
}

func (m *customVideoTagModel) withSession(session sqlx.Session) VideoTagModel {
	return NewVideoTagModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customVideoTagModel) InsertBatch(ctx context.Context, list []*VideoTag) error {
	// 构造 query
	if len(list) == 0 {
		return nil
	}

	valuePlaceholders := make([]string, 0, len(list)) // 生成 占位符
	valueArgs := make([]interface{}, 0, len(list)*3)  // 存放 插入的参数

	for _, tag := range list {
		valuePlaceholders = append(valuePlaceholders, "(?,?,?)")
		valueArgs = append(valueArgs, tag.VideoID, tag.Tag, tag.DeletedAt)
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES %s`,
		m.table,
		videoTagRowsExpectAutoSet,
		strings.Join(valuePlaceholders, ","),
	)
	_, err := m.conn.ExecCtx(ctx, query, valueArgs...)
	return mapDBError(err)
}

func (m *customVideoTagModel) InsertBatchWithSession(ctx context.Context, session sqlx.Session, list []*VideoTag) error {
	return m.withSession(session).InsertBatch(ctx, list)
}

func (m *customVideoTagModel) FindByVideoTags(ctx context.Context, offest, limit int, tags []string) ([]*VideoTag, error) {
	// 为了安全 要使用占位符
	placeholders := make([]string, 0, len(tags))
	valueArgs := make([]interface{}, 0, len(tags)+2) // 占位符中的数据，后两个是 offest 和 limit

	for _, tag := range tags {
		placeholders = append(placeholders, "?")
		valueArgs = append(valueArgs, tag)
	}
	valueArgs = append(valueArgs, offest, limit)

	query := fmt.Sprintf("select %s from %s where `tag` in (%s) limit ?,?",
		videoTagRows,
		m.table,
		strings.Join(placeholders, ","),
	)
	var out []*VideoTag
	err := m.conn.QueryRowsCtx(ctx, &out, query, valueArgs...)
	return out, mapDBError(err)
}

func (m *customVideoTagModel) FindManyByVideoID(ctx context.Context, videoID uint64) ([]*VideoTag, error) {
	query := fmt.Sprintf("select %s from %s where `video_id` = ? and `deleted_at` = 0", videoTagRows, m.table)
	var res []*VideoTag
	err := m.conn.QueryRowsCtx(ctx, &res, query, videoID)
	return res, mapDBError(err)
}

func (m *customVideoTagModel) DeleteBatchByVideoID(ctx context.Context, videoID uint64) error {
	query := fmt.Sprintf("delete from %s where `video_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, videoID)
	return mapDBError(err)
}

func (m *customVideoTagModel) DeleteBatchByVideoIDWithSession(ctx context.Context, session sqlx.Session, videoID uint64) error {
	return m.withSession(session).DeleteBatchByVideoID(ctx, videoID)
}

func (m *customVideoTagModel) SoftDeleteByVideoID(ctx context.Context, videoID, deletedAt uint64) error {
	query := fmt.Sprintf(
		"update %s set `deleted_at` = ? where `video_id` = ?",
		m.table,
	)

	_, err := m.conn.ExecCtx(ctx, query, deletedAt, videoID)
	return mapDBError(err)
}

func (m *customVideoTagModel) SoftDeleteByVideoIDWithSession(ctx context.Context, session sqlx.Session, videoID, deletedAt uint64) error {
	return m.withSession(session).SoftDeleteByVideoID(ctx, videoID, deletedAt)
}
