package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ComicModel = (*customComicModel)(nil)

const (
	comicTagTable = "comic_tag"
)

type (
	// ComicModel is an interface to be customized, add more methods here,
	// and implement the added methods in customComicModel.
	ComicModel interface {
		comicModel
		withSession(session sqlx.Session) ComicModel
		FindByTagsAndKeyWord(ctx context.Context, offset int, limit int, tags []string, keyword string) ([]*Comic, error)
	}

	customComicModel struct {
		*defaultComicModel
	}
)

// NewComicModel returns a model for the database table.
func NewComicModel(conn sqlx.SqlConn) ComicModel {
	return &customComicModel{
		defaultComicModel: newComicModel(conn),
	}
}

func (m *customComicModel) withSession(session sqlx.Session) ComicModel {
	return NewComicModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customComicModel) FindByTagsAndKeyWord(ctx context.Context, offset int, limit int, tags []string, keyword string) ([]*Comic, error) {
	kw := "%" + keyword + "%"
	args := make([]interface{}, 0, len(tags)+5) // 占位符的数据，后两个是 offest 和 limit
	args = append(args, kw, kw, kw)             // name, description, author
	tagFilter := ""                             // tags 为空时跳过筛选
	if len(tags) > 0 {
		placeholders := make([]string, 0, len(tags))
		for _, tag := range tags {
			placeholders = append(placeholders, "?")
			args = append(args, tag)
		}
		tagFilter = "and t.tag in (" + strings.Join(placeholders, ",") + ")"
	}
	args = append(args, offset, limit)

	// distinct 去重
	query := fmt.Sprintf(`
		select distinct c.*
		from %s c
		left join %s t on c.id = t.comic_id
		where (c.name like ? or c.description like ? or c.author like ?)
		%s
		limit ?,?
`, m.table, comicTagTable, tagFilter)

	var out []*Comic
	err := m.conn.QueryRowsCtx(ctx, &out, query, args...)
	return out, err
}
