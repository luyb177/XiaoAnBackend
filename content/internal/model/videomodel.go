package model

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
)

var _ VideoModel = (*customVideoModel)(nil)

const (
	videoTagTable = "video_tag"
)

type (
	// VideoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customVideoModel.
	VideoModel interface {
		videoModel
		withSession(session sqlx.Session) VideoModel
		InsertWithSession(ctx context.Context, session sqlx.Session, data *Video) (sql.Result, error)
		FindByKeyWord(ctx context.Context, offset int, limit int, keyword string) ([]*Video, error)
		FindByVideoTagsAndKeyWord(ctx context.Context, offset int, limit int, tags []string, keyword string) ([]*Video, error)
	}

	customVideoModel struct {
		*defaultVideoModel
	}
)

// NewVideoModel returns a model for the database table.
func NewVideoModel(conn sqlx.SqlConn) VideoModel {
	return &customVideoModel{
		defaultVideoModel: newVideoModel(conn),
	}
}

func (m *customVideoModel) withSession(session sqlx.Session) VideoModel {
	return NewVideoModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customVideoModel) InsertWithSession(ctx context.Context, session sqlx.Session, data *Video) (sql.Result, error) {
	return m.withSession(session).Insert(ctx, data)
}

func (m *customVideoModel) FindByKeyWord(ctx context.Context, offset int, limit int, keyword string) ([]*Video, error) {
	kw := "%" + keyword + "%"
	query := fmt.Sprintf("select %s from %s where name like ? or description like ? or author like ? limit ?, ?", videoRows, m.table)
	var out []*Video
	err := m.conn.QueryRowsCtx(ctx, &out, query, kw, kw, kw, offset, limit)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m *customVideoModel) FindByVideoTagsAndKeyWord(ctx context.Context, offset int, limit int, tags []string, keyword string) ([]*Video, error) {
	kw := "%" + keyword + "%"

	args := make([]interface{}, 0, len(tags)+5) // 占位符的数据，后两个是 offest 和 limit
	args = append(args, kw, kw, kw)             // name, description, author

	tagFilter := "" // tags 为空时跳过筛选
	if len(tags) > 0 {
		placeholders := make([]string, 0, len(tags))
		for _, tag := range tags {
			placeholders = append(placeholders, "?")
			args = append(args, tag)
		}
		tagFilter = "and t.tag in (" + strings.Join(placeholders, ",") + ")"
	}

	args = append(args, offset, limit)

	quary := fmt.Sprintf(`
		select distinct v.*
		from %s v
		left join %s t on v.id = t.video_id
		where (v.name like ? or v.description like ? or v.author like ?)
		%s
		limit ?,?
`, m.table, videoTagTable, tagFilter)

	var out []*Video
	err := m.conn.QueryRowsCtx(ctx, &out, quary, args...)
	return out, err
}
