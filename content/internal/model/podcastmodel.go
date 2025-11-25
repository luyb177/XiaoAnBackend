package model

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
)

var _ PodcastModel = (*customPodcastModel)(nil)

const (
	podcastTagTable = "podcast_tag"
)

type (
	// PodcastModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPodcastModel.
	PodcastModel interface {
		podcastModel
		withSession(session sqlx.Session) PodcastModel
		FindByTagsAndKeyWord(ctx context.Context, offset int, limit int, tags []string, keyword string) ([]*Podcast, error)
	}

	customPodcastModel struct {
		*defaultPodcastModel
	}
)

// NewPodcastModel returns a model for the database table.
func NewPodcastModel(conn sqlx.SqlConn) PodcastModel {
	return &customPodcastModel{
		defaultPodcastModel: newPodcastModel(conn),
	}
}

func (m *customPodcastModel) withSession(session sqlx.Session) PodcastModel {
	return NewPodcastModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customPodcastModel) FindByTagsAndKeyWord(ctx context.Context, offset int, limit int, tags []string, keyword string) ([]*Podcast, error) {
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

	// distinct 去重
	query := fmt.Sprintf(`
		select distinct p.*
		from %s p
		left join %s t on p.id = t.podcast_id
		where (p.name like ? or p.description like ? or p.author like ?)
		%s
		limit ?,?
`, m.table, podcastTagTable, tagFilter)

	var out []*Podcast
	err := m.conn.QueryRowsCtx(ctx, &out, query, args...)
	return out, err
}
