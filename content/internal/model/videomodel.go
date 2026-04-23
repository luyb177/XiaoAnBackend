package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
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
		IncrCommentCount(ctx context.Context, videoID uint64) (sql.Result, error)
		IncrCommentCountWithSession(ctx context.Context, session sqlx.Session, videoID uint64) (sql.Result, error)
		DecrCommentCount(ctx context.Context, videoID uint64) (sql.Result, error)
		DecrCommentCountWithSession(ctx context.Context, session sqlx.Session, videoID uint64) (sql.Result, error)
		DecrCommentCountByCount(ctx context.Context, videoID, count uint64) (sql.Result, error)
		DecrCommentCountByCountWithSession(ctx context.Context, session sqlx.Session, videoID, count uint64) (sql.Result, error)
		IncrLikeCount(ctx context.Context, videoID uint64) (sql.Result, error)
		IncrLikeCountWithSession(ctx context.Context, session sqlx.Session, videoID uint64) (sql.Result, error)
		DecrLikeCount(ctx context.Context, id uint64) (sql.Result, error)
		DecrLikeCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
		IncrViewCount(ctx context.Context, videoID uint64) (sql.Result, error)
		IncrViewCountWithSession(ctx context.Context, session sqlx.Session, videoID uint64) (sql.Result, error)
		IncrCollectCount(ctx context.Context, id uint64) (sql.Result, error)
		IncrCollectCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
		DecrCollectCount(ctx context.Context, id uint64) (sql.Result, error)
		DecrCollectCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
		FindManyWithNotDelete(ctx context.Context, limit int64) ([]*Video, error)
		FindManyWithNotDeleteByCursor(ctx context.Context, cursor uint64, limit int64) ([]*Video, error)
		FindByKeyWord(ctx context.Context, offset, limit int, keyword string) ([]*Video, error)
		FindByVideoTagsAndKeyWord(ctx context.Context, offset, limit int, tags []string, keyword string) ([]*Video, error)
		FindOneWithNotDelete(ctx context.Context, id uint64) (*Video, error)
		FindOneWithNotDeleteWithSession(ctx context.Context, session sqlx.Session, id uint64) (*Video, error)
		UpdateRelationStatus(ctx context.Context, id uint64, relationStatus int64) error
		UpdateRelationStatusWithSession(ctx context.Context, session sqlx.Session, id uint64, relationStatus int64) error
		SoftDelete(ctx context.Context, id uint64, deletedAt uint64, modifier sql.NullInt64) error
		UpdateAuthorByLastModifiedBy(ctx context.Context, lastModifiedBy int64, author string) error
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

func (m *customVideoModel) IncrCommentCount(ctx context.Context, videoID uint64) (sql.Result, error) {
	query := fmt.Sprintf("update %s set `comment_count` = `comment_count` + 1 where `id` = ? and `deleted_at` = 0", m.table)

	result, err := m.conn.ExecCtx(ctx, query, videoID)
	return result, mapDBError(err)
}

func (m *customVideoModel) IncrCommentCountWithSession(ctx context.Context, session sqlx.Session, videoID uint64) (sql.Result, error) {
	return m.withSession(session).IncrCommentCount(ctx, videoID)
}

func (m *customVideoModel) DecrCommentCount(ctx context.Context, videoID uint64) (sql.Result, error) {
	query := fmt.Sprintf("update %s set `comment_count` = `comment_count` - 1 where `id` = ? and `comment_count` > 0 and `deleted_at` =0", m.table)

	result, err := m.conn.ExecCtx(ctx, query, videoID)
	return result, mapDBError(err)
}

func (m *customVideoModel) DecrCommentCountWithSession(ctx context.Context, session sqlx.Session, videoID uint64) (sql.Result, error) {
	return m.withSession(session).DecrCommentCount(ctx, videoID)
}

func (m *customVideoModel) DecrCommentCountByCount(ctx context.Context, videoID, count uint64) (sql.Result, error) {
	query := fmt.Sprintf("update %s set `comment_count` = `comment_count` - ? where `id` = ? and `comment_count` >= ? and `deleted_at` =0", m.table)

	result, err := m.conn.ExecCtx(ctx, query, count, videoID, count)
	return result, mapDBError(err)
}

func (m *customVideoModel) DecrCommentCountByCountWithSession(ctx context.Context, session sqlx.Session, videoID, count uint64) (sql.Result, error) {
	return m.withSession(session).DecrCommentCountByCount(ctx, videoID, count)
}

func (m *customVideoModel) IncrLikeCount(ctx context.Context, videoID uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set like_count = like_count + 1
		where id = ? and deleted_at = 0`,
		m.table)

	result, err := m.conn.ExecCtx(ctx, query, videoID)
	return result, mapDBError(err)
}

func (m *customVideoModel) IncrLikeCountWithSession(ctx context.Context, session sqlx.Session, videoID uint64) (sql.Result, error) {
	return m.withSession(session).IncrLikeCount(ctx, videoID)
}

func (m *customVideoModel) DecrLikeCount(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set like_count = like_count - 1
		where id = ? and like_count > 0 and deleted_at = 0`,
		m.table,
	)

	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customVideoModel) DecrLikeCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).DecrLikeCount(ctx, id)
}

func (m *customVideoModel) IncrViewCount(ctx context.Context, videoID uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set view_count = view_count + 1
		where id = ? and deleted_at = 0`,
		m.table,
	)

	result, err := m.conn.ExecCtx(ctx, query, videoID)
	return result, mapDBError(err)
}

func (m *customVideoModel) IncrViewCountWithSession(ctx context.Context, session sqlx.Session, videoID uint64) (sql.Result, error) {
	return m.withSession(session).IncrViewCount(ctx, videoID)
}

func (m *customVideoModel) IncrCollectCount(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set collect_count = collect_count + 1
		where id = ? and deleted_at = 0`,
		m.table,
	)

	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customVideoModel) IncrCollectCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).IncrCollectCount(ctx, id)
}

func (m *customVideoModel) DecrCollectCount(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s
		set collect_count = collect_count - 1
		where id = ? and collect_count > 0 and deleted_at = 0`,
		m.table,
	)

	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customVideoModel) DecrCollectCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).DecrCollectCount(ctx, id)
}

func (m *customVideoModel) FindOneWithNotDelete(ctx context.Context, id uint64) (*Video, error) {
	query := fmt.Sprintf(
		"select %s from %s where `id` = ? and `deleted_at` = 0 limit 1",
		videoRows,
		m.table,
	)

	var resp Video
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	return &resp, mapDBError(err)
}

func (m *customVideoModel) FindOneWithNotDeleteWithSession(ctx context.Context, session sqlx.Session, id uint64) (*Video, error) {
	return m.withSession(session).FindOneWithNotDelete(ctx, id)
}

func (m *customVideoModel) FindManyWithNotDelete(ctx context.Context, limit int64) ([]*Video, error) {
	query := fmt.Sprintf(`
		select %s from %s
		where deleted_at = 0
		order by id desc
		limit ?`,
		videoRows,
		m.table,
	)

	var out []*Video
	err := m.conn.QueryRowsCtx(ctx, &out, query, limit)
	return out, mapDBError(err)
}

func (m *customVideoModel) FindManyWithNotDeleteByCursor(ctx context.Context, cursor uint64, limit int64) ([]*Video, error) {
	query := fmt.Sprintf(`
		select %s from %s
		where deleted_at = 0 
		    and id < ?
		order by id desc
		limit ?`,
		videoRows,
		m.table,
	)

	var out []*Video
	err := m.conn.QueryRowsCtx(ctx, &out, query, cursor, limit)
	return out, mapDBError(err)
}

func (m *customVideoModel) FindByKeyWord(ctx context.Context, offset, limit int, keyword string) ([]*Video, error) {
	kw := "%" + keyword + "%"
	query := fmt.Sprintf("select %s from %s where name like ? or description like ? or author like ? limit ?, ?", videoRows, m.table)
	var out []*Video
	err := m.conn.QueryRowsCtx(ctx, &out, query, kw, kw, kw, offset, limit)
	return out, mapDBError(err)
}

func (m *customVideoModel) FindByVideoTagsAndKeyWord(ctx context.Context, offset, limit int, tags []string, keyword string) ([]*Video, error) {
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
	return out, mapDBError(err)
}

func (m *customVideoModel) UpdateRelationStatus(ctx context.Context, id uint64, relationStatus int64) error {
	query := fmt.Sprintf("update %s set `relation_status` = ? where `id` = ?", m.table)

	_, err := m.conn.ExecCtx(ctx, query, relationStatus, id)
	return mapDBError(err)
}

func (m *customVideoModel) UpdateRelationStatusWithSession(ctx context.Context, session sqlx.Session, id uint64, relationStatus int64) error {
	return m.withSession(session).UpdateRelationStatus(ctx, id, relationStatus)
}

func (m *customVideoModel) SoftDelete(ctx context.Context, id, deletedAt uint64, modifier sql.NullInt64) error {
	query := fmt.Sprintf(
		"update %s set `deleted_at` = ?, `last_modified_by` = ? where `id` = ?",
		m.table,
	)

	_, err := m.conn.ExecCtx(ctx, query, deletedAt, modifier, id)
	return mapDBError(err)
}

func (m *customVideoModel) UpdateAuthorByLastModifiedBy(ctx context.Context, lastModifiedBy int64, author string) error {
	query := fmt.Sprintf("update %s set `author` = ? where `last_modified_by` = ? and `deleted_at` = 0", m.table)
	_, err := m.conn.ExecCtx(ctx, query, author, lastModifiedBy)
	return mapDBError(err)
}
