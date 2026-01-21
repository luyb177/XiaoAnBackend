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
		FindManyByVideoId(ctx context.Context, videoId uint64) ([]*VideoTag, error)
		DeleteBatchByVideoId(ctx context.Context, videoId uint64) error
		DeleteBatchByVideoIdWithSession(ctx context.Context, session sqlx.Session, videoId uint64) error
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

	for _, v := range list {
		valuePlaceholders = append(valuePlaceholders, "(?,?,?)")
		valueArgs = append(valueArgs, v.VideoId, v.Tag, v.DeletedAt)
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES %s`,
		m.table,
		videoTagRowsExpectAutoSet,
		strings.Join(valuePlaceholders, ","),
	)
	_, err := m.conn.ExecCtx(ctx, query, valueArgs...)
	return err
}

func (m *customVideoTagModel) InsertBatchWithSession(ctx context.Context, session sqlx.Session, list []*VideoTag) error {
	return m.withSession(session).InsertBatch(ctx, list)
}

func (m *customVideoTagModel) FindByVideoTags(ctx context.Context, offest int, limit int, tags []string) ([]*VideoTag, error) {
	// 为了安全 要使用占位符
	placeholders := make([]string, 0, len(tags))
	valueArgs := make([]interface{}, 0, len(tags)+2) // 占位符中的数据，后两个是 offest 和 limit

	for _, v := range tags {
		placeholders = append(placeholders, "?")
		valueArgs = append(valueArgs, v)
	}
	valueArgs = append(valueArgs, offest, limit)

	query := fmt.Sprintf("select %s from %s where `tag` in (%s) limit ?,?",
		videoTagRows,
		m.table,
		strings.Join(placeholders, ","),
	)
	var out []*VideoTag
	err := m.conn.QueryRowsCtx(ctx, &out, query, valueArgs...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m *customVideoTagModel) FindManyByVideoId(ctx context.Context, videoId uint64) ([]*VideoTag, error) {
	query := fmt.Sprintf("select %s from %s where `video_id` = ?", videoTagRows, m.table)
	var res []*VideoTag
	err := m.conn.QueryRowsCtx(ctx, &res, query, videoId)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *customVideoTagModel) DeleteBatchByVideoId(ctx context.Context, videoId uint64) error {
	query := fmt.Sprintf("delete from %s where `video_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, videoId)
	return err
}

func (m *customVideoTagModel) DeleteBatchByVideoIdWithSession(ctx context.Context, session sqlx.Session, videoId uint64) error {
	return m.withSession(session).DeleteBatchByVideoId(ctx, videoId)
}
