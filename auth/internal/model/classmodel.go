package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ClassModel = (*customClassModel)(nil)

type (
	// ClassModel is an interface to be customized, add more methods here,
	// and implement the added methods in customClassModel.
	ClassModel interface {
		classModel
		withSession(session sqlx.Session) ClassModel
		Insert(ctx context.Context, data *Class) (sql.Result, error)
		FindNormalOne(ctx context.Context, id uint64) (*Class, error)
		FindManyByUserID(ctx context.Context, userID uint64, limit int64) ([]*Class, error)
		FindManyByUserIDWithCursor(ctx context.Context, userID, cursor uint64, limit int64) ([]*Class, error)
		IncrStudentCount(ctx context.Context, id uint64) (sql.Result, error)
		IncrStudentCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
		DecrStudentCount(ctx context.Context, id uint64) (sql.Result, error)
		DecrStudentCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error)
	}

	customClassModel struct {
		*defaultClassModel
	}
)

// NewClassModel returns a model for the database table.
func NewClassModel(conn sqlx.SqlConn) ClassModel {
	return &customClassModel{
		defaultClassModel: newClassModel(conn),
	}
}

func (m *customClassModel) withSession(session sqlx.Session) ClassModel {
	return NewClassModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customClassModel) Insert(ctx context.Context, data *Class) (sql.Result, error) {
	//nolint:staticcheck // QF1008: go-zero embedding style retained intentionally
	result, err := m.defaultClassModel.Insert(ctx, data)
	return result, mapDBError(err)
}

func (m *customClassModel) FindNormalOne(ctx context.Context, id uint64) (*Class, error) {
	query := fmt.Sprintf(`
		select %s from %s
		where id = ? and deleted_at = 0 and status = 1`,
		classRows,
		m.table,
	)

	var resp Class
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	return &resp, mapDBError(err)
}

func (m *customClassModel) FindManyByUserID(ctx context.Context, userID uint64, limit int64) ([]*Class, error) {
	query := fmt.Sprintf(`
		select %s from %s
		where admin_id = ? 
		    and deleted_at = 0 
		    and status = 1
		order by id desc
		limit ?`,
		classRows,
		m.table,
	)

	var resp []*Class
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userID, limit)
	return resp, mapDBError(err)
}

func (m *customClassModel) FindManyByUserIDWithCursor(ctx context.Context, userID, cursor uint64, limit int64) ([]*Class, error) {
	query := fmt.Sprintf(`
		select %s from %s
		where admin_id = ? 
		    and deleted_at = 0 
		    and status = 1
		    and id < ?
		order by id desc
		limit ?`,
		classRows,
		m.table,
	)

	var resp []*Class
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userID, cursor, limit)
	return resp, mapDBError(err)
}

func (m *customClassModel) IncrStudentCount(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s set student_count = student_count + 1
		where id = ? and deleted_at = 0 and status = 1`,
		m.table)

	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customClassModel) IncrStudentCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).IncrStudentCount(ctx, id)
}

func (m *customClassModel) DecrStudentCount(ctx context.Context, id uint64) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s set student_count = student_count - 1
		where id = ? and deleted_at = 0 and status = 1 and student_count > 0`,
		m.table)

	result, err := m.conn.ExecCtx(ctx, query, id)
	return result, mapDBError(err)
}

func (m *customClassModel) DecrStudentCountWithSession(ctx context.Context, session sqlx.Session, id uint64) (sql.Result, error) {
	return m.withSession(session).DecrStudentCount(ctx, id)
}
