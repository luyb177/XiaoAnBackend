package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ InviteCodeModel = (*customInviteCodeModel)(nil)

type (
	// InviteCodeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customInviteCodeModel.
	InviteCodeModel interface {
		inviteCodeModel
		withSession(session sqlx.Session) InviteCodeModel
		Insert(ctx context.Context, data *InviteCode) (sql.Result, error)
		FindUsableByCode(ctx context.Context, code string) (*InviteCode, error)
		FindOneByCodeWithNotDelete(ctx context.Context, code string) (*InviteCode, error)
		FindManyByCreatorID(ctx context.Context, creatorID uint64, pageSize int64) ([]*InviteCode, error)
		FindManyByCreatorIDWithCursor(ctx context.Context, creatorID, cursor uint64, pageSize int64) ([]*InviteCode, error)
		CountByCreatorID(ctx context.Context, creatorID uint64) (int64, error)
		Update(ctx context.Context, data *InviteCode) error
		UpdateWithSession(ctx context.Context, session sqlx.Session, data *InviteCode) error
		IncrUsedCount(ctx context.Context, id uint64) (sql.Result, error)
		IncrUsedCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
	}

	customInviteCodeModel struct {
		*defaultInviteCodeModel
	}
)

// NewInviteCodeModel returns a model for the database table.
func NewInviteCodeModel(conn sqlx.SqlConn) InviteCodeModel {
	return &customInviteCodeModel{
		defaultInviteCodeModel: newInviteCodeModel(conn),
	}
}

func (m *customInviteCodeModel) withSession(session sqlx.Session) InviteCodeModel {
	return NewInviteCodeModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customInviteCodeModel) Insert(ctx context.Context, data *InviteCode) (sql.Result, error) {
	//nolint:staticcheck // QF1008: go-zero embedding style retained intentionally
	result, err := m.defaultInviteCodeModel.Insert(ctx, data)
	return result, mapDBError(err)
}

func (m *customInviteCodeModel) FindUsableByCode(ctx context.Context, code string) (*InviteCode, error) {
	query := fmt.Sprintf(`
		select %s from %s
		where code = ? 
			and is_active = 1 
			and deleted_at = 0 
			and max_uses > used_count
			and expires_at > now()`,
		inviteCodeRows,
		m.table)

	var resp InviteCode
	err := m.conn.QueryRowCtx(ctx, &resp, query, code)
	return &resp, mapDBError(err)
}

func (m *customInviteCodeModel) FindOneByCodeWithNotDelete(ctx context.Context, code string) (*InviteCode, error) {
	query := fmt.Sprintf(`
		select %s from %s
		where code = ? and deleted_at = 0`,
		inviteCodeRows,
		m.table)

	var resp InviteCode
	err := m.conn.QueryRowCtx(ctx, &resp, query, code)
	return &resp, mapDBError(err)
}

// FindManyByCreatorID 根据创建者ID分页查询邀请码列表
func (m *customInviteCodeModel) FindManyByCreatorID(ctx context.Context, creatorID uint64, pageSize int64) ([]*InviteCode, error) {
	query := fmt.Sprintf(`
		select %s from %s
		where creator_id = ? 
		    and deleted_at = 0 
		order by id desc
		limit ?`,
		inviteCodeRows,
		m.table,
	)

	var resp []*InviteCode
	err := m.conn.QueryRowsCtx(ctx, &resp, query, creatorID, pageSize)
	return resp, mapDBError(err)
}

func (m *customInviteCodeModel) FindManyByCreatorIDWithCursor(ctx context.Context, creatorID, cursor uint64, pageSize int64) ([]*InviteCode, error) {
	query := fmt.Sprintf(`
		select %s from %s
		where creator_id = ? 
		    and deleted_at = 0 
		    and id < ? 
		order by id desc
		limit ?`,
		inviteCodeRows,
		m.table,
	)

	var resp []*InviteCode
	err := m.conn.QueryRowsCtx(ctx, &resp, query, creatorID, cursor, pageSize)
	return resp, mapDBError(err)
}

// CountByCreatorID 统计创建者的邀请码总数
func (m *customInviteCodeModel) CountByCreatorID(ctx context.Context, creatorID uint64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where `creator_id` = ?", m.table)

	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, creatorID)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (m *customInviteCodeModel) Update(ctx context.Context, data *InviteCode) error {
	//nolint:staticcheck // QF1008: go-zero embedding style retained intentionally
	err := m.defaultInviteCodeModel.Update(ctx, data)
	return mapDBError(err)
}

func (m *customInviteCodeModel) UpdateWithSession(ctx context.Context, session sqlx.Session, data *InviteCode) error {
	return m.withSession(session).Update(ctx, data)
}

func (m *customInviteCodeModel) IncrUsedCount(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s 
		set used_count = used_count + 1,
		    is_active = case 
				when used_count + 1 >= max_uses then 0
				else is_active
			end
		where id = ? 
			and deleted_at = 0
			and is_active = 1
			and used_count < max_uses`,
		m.table,
	)
	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customInviteCodeModel) IncrUsedCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).IncrUsedCount(ctx, id)
}
