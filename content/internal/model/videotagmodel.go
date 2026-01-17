package model

import (
	"context"
	"database/sql"
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
		BatchInsertWithSession(ctx context.Context, session sqlx.Session, data []VideoTag) (sql.Result, error)
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

func (m *customVideoTagModel) BatchInsertWithSession(ctx context.Context, session sqlx.Session, data []VideoTag) (sql.Result, error) {
	// 构造 query
	if len(data) == 0 {
		return nil, nil
	}
	placeholders := make([]string, 0, len(data))     // 生成 占位符
	valueArgs := make([]interface{}, 0, len(data)*2) // 存放 插入的参数
	for _, v := range data {
		placeholders = append(placeholders, "(?,?)")
		valueArgs = append(valueArgs, v.VideoId, v.Tag)
	}
	query := fmt.Sprintf("insert into %s (%s) values %s", m.table, videoTagRowsWithPlaceHolder, strings.Join(placeholders, ","))
	return session.ExecCtx(ctx, query, valueArgs...)
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

func (m *customVideoTagModel) FindByVideoId(ctx context.Context, videoId uint64) ([]*VideoTag, error) {
	query := fmt.Sprintf("select %s from %s where `video_id` = ?", videoTagRows, m.table)
	var out []*VideoTag
	err := m.conn.QueryRowsCtx(ctx, &out, query, videoId)
	if err != nil {
		return nil, err
	}
	return out, nil
}
