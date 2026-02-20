package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
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
		IncrCommentCount(ctx context.Context, podcastID uint64) (sql.Result, error)
		IncrCommentCountWithSession(ctx context.Context, session sqlx.Session, podcastID uint64) (sql.Result, error)
		DecrCommentCount(ctx context.Context, podcastID uint64) (sql.Result, error)
		DecrCommentCountWithSession(ctx context.Context, session sqlx.Session, podcastID uint64) (sql.Result, error)
		DecrCommentCountByCount(ctx context.Context, podcastID, count uint64) (sql.Result, error)
		DecrCommentCountByCountWithSession(ctx context.Context, session sqlx.Session, podcastID, count uint64) (sql.Result, error)
		IncrLikeCount(ctx context.Context, podcastID uint64) (sql.Result, error)
		IncrLikeCountWithSession(ctx context.Context, session sqlx.Session, podcastID uint64) (sql.Result, error)
		DecrLikeCount(ctx context.Context, id uint64) (sql.Result, error)
		DecrLikeCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
		IncrViewCount(ctx context.Context, id uint64) (sql.Result, error)
		IncrViewCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
		IncrCollectCount(ctx context.Context, id uint64) (sql.Result, error)
		IncrCollectCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
		DecrCollectCount(ctx context.Context, id uint64) (sql.Result, error)
		DecrCollectCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
		FindOneWithNotDelete(ctx context.Context, id uint64) (*Podcast, error)
		FindOneWithNotDeleteWithSession(ctx context.Context, session sqlx.Session, id uint64) (*Podcast, error)
		FindByTagsAndKeyWord(ctx context.Context, offset, limit int, tags []string, keyword string) ([]*Podcast, error)
		UpdateRelationStatus(ctx context.Context, id uint64, relationStatus int64) error
		UpdateRelationStatusWithSession(ctx context.Context, session sqlx.Session, id uint64, relationStatus int64) error
		SoftDelete(ctx context.Context, id uint64, deletedAt uint64, modifier sql.NullInt64) error
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

func (m *customPodcastModel) IncrCommentCount(ctx context.Context, podcastID uint64) (sql.Result, error) {
	query := fmt.Sprintf("update %s set `comment_count` = `comment_count` + 1 where `id` = ? and `deleted_at` = 0", m.table)
	result, err := m.conn.ExecCtx(ctx, query, podcastID)
	return result, mapDBError(err)
}

func (m *customPodcastModel) IncrCommentCountWithSession(ctx context.Context, session sqlx.Session, podcastID uint64) (sql.Result, error) {
	return m.withSession(session).IncrCommentCount(ctx, podcastID)
}

func (m *customPodcastModel) DecrCommentCount(ctx context.Context, podcastID uint64) (sql.Result, error) {
	query := fmt.Sprintf("update %s set `comment_count` = `comment_count` - 1 where `id` = ? and `comment_count` > 0 and `deleted_at` = 0", m.table)
	result, err := m.conn.ExecCtx(ctx, query, podcastID)
	return result, mapDBError(err)
}

func (m *customPodcastModel) DecrCommentCountWithSession(ctx context.Context, session sqlx.Session, podcastID uint64) (sql.Result, error) {
	return m.withSession(session).DecrCommentCount(ctx, podcastID)
}

func (m *customPodcastModel) DecrCommentCountByCount(ctx context.Context, podcastID, count uint64) (sql.Result, error) {
	query := fmt.Sprintf("update %s set `comment_count` = `comment_count` - ? where `id` = ? and `comment_count` >= ? and `deleted_at` = 0", m.table)
	result, err := m.conn.ExecCtx(ctx, query, count, podcastID, count)
	return result, mapDBError(err)
}

func (m *customPodcastModel) DecrCommentCountByCountWithSession(ctx context.Context, session sqlx.Session, podcastID, count uint64) (sql.Result, error) {
	return m.withSession(session).DecrCommentCountByCount(ctx, podcastID, count)
}

func (m *customPodcastModel) IncrLikeCount(ctx context.Context, podcastID uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set like_count = like_count + 1
		where id = ? and deleted_at = 0`,
		m.table,
	)
	result, err := m.conn.ExecCtx(ctx, query, podcastID)
	return result, mapDBError(err)
}

func (m *customPodcastModel) IncrLikeCountWithSession(ctx context.Context, session sqlx.Session, podcastID uint64) (sql.Result, error) {
	return m.withSession(session).IncrLikeCount(ctx, podcastID)
}

func (m *customPodcastModel) DecrLikeCount(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set like_count = like_count - 1
		where id = ? and like_count > 0 and deleted_at = 0`,
		m.table,
	)

	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customPodcastModel) DecrLikeCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).DecrLikeCount(ctx, id)
}

func (m *customPodcastModel) IncrViewCount(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set view_count = view_count + 1
		where id = ? and deleted_at = 0`,
		m.table,
	)

	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customPodcastModel) IncrViewCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).IncrViewCount(ctx, id)
}

func (m *customPodcastModel) IncrCollectCount(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set collect_count = collect_count + 1
		where id = ? and deleted_at = 0`,
		m.table,
	)

	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customPodcastModel) IncrCollectCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).IncrCollectCount(ctx, id)
}

func (m *customPodcastModel) DecrCollectCount(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set collect_count = collect_count - 1
		where id = ? and collect_count > 0 and deleted_at = 0`,
		m.table,
	)

	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customPodcastModel) DecrCollectCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).DecrCollectCount(ctx, id)
}

func (m *customPodcastModel) FindByTagsAndKeyWord(ctx context.Context, offset, limit int, tags []string, keyword string) ([]*Podcast, error) {
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
	return out, mapDBError(err)
}

func (m *customPodcastModel) UpdateRelationStatus(ctx context.Context, id uint64, relationStatus int64) error {
	query := fmt.Sprintf("update %s set `relation_status` = ? where `id` = ?", m.table)

	_, err := m.conn.ExecCtx(ctx, query, relationStatus, id)
	return mapDBError(err)
}

func (m *customPodcastModel) UpdateRelationStatusWithSession(ctx context.Context, session sqlx.Session, id uint64, relationStatus int64) error {
	return m.withSession(session).UpdateRelationStatus(ctx, id, relationStatus)
}

func (m *customPodcastModel) FindOneWithNotDelete(ctx context.Context, id uint64) (*Podcast, error) {
	query := fmt.Sprintf(
		"select %s from %s where id = ? and `deleted_at` = 0 limit 1",
		podcastRows,
		m.table,
	)

	var resp Podcast
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	return &resp, mapDBError(err)
}

func (m *customPodcastModel) FindOneWithNotDeleteWithSession(ctx context.Context, session sqlx.Session, id uint64) (*Podcast, error) {
	return m.withSession(session).FindOneWithNotDelete(ctx, id)
}

func (m *customPodcastModel) SoftDelete(ctx context.Context, id, deletedAt uint64, modifier sql.NullInt64) error {
	query := fmt.Sprintf(
		"update %s set `deleted_at` = ?, `last_modified_by` = ? where `id` = ?",
		m.table,
	)
	_, err := m.conn.ExecCtx(ctx, query, deletedAt, modifier, id)
	return mapDBError(err)
}
