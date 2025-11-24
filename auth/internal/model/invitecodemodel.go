package model

import (
	"context"
	"errors"
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
		FindByCreatorId(ctx context.Context, creatorId uint64, page, pageSize int64) ([]*InviteCode, error)
		CountByCreatorId(ctx context.Context, creatorId uint64) (int64, error)
		BatchInsert(ctx context.Context, datas []InviteCode) error
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

// FindByCreatorId 根据创建者ID分页查询邀请码列表
func (m *customInviteCodeModel) FindByCreatorId(ctx context.Context, creatorId uint64, page, pageSize int64) ([]*InviteCode, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("select %s from %s where `creator_id` = ? order by `created_at` desc limit ? offset ?", inviteCodeRows, m.table)

	var resp []*InviteCode
	err := m.conn.QueryRowsCtx(ctx, &resp, query, creatorId, pageSize, offset)
	switch {
	case err == nil:
		return resp, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return []*InviteCode{}, nil
	default:
		return nil, err
	}
}

// CountByCreatorId 统计创建者的邀请码总数
func (m *customInviteCodeModel) CountByCreatorId(ctx context.Context, creatorId uint64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where `creator_id` = ?", m.table)

	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, creatorId)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// BatchInsert 事务 生成大量数据
func (m *customInviteCodeModel) BatchInsert(ctx context.Context, datas []InviteCode) error {
	return m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, inviteCodeRowsExpectAutoSet)

		for _, data := range datas {
			_, err := session.ExecCtx(ctx, query, data.Code, data.CreatorId, data.CreatorName, data.Department, data.MaxUses, data.UsedCount, data.IsActive, data.Remark, data.ExpiresAt, data.TargetRole, data.ClassId, data.Type)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
