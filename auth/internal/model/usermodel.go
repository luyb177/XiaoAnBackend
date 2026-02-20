package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserModel = (*customUserModel)(nil)

type (
	// UserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserModel.
	UserModel interface {
		userModel
		withSession(session sqlx.Session) UserModel
		Insert(ctx context.Context, data *User) (sql.Result, error)
		InsertWithSession(ctx context.Context, session sqlx.Session, data *User) (sql.Result, error)
		FindOneWithNotDelete(ctx context.Context, id uint64) (*User, error)
		FindOneByEmailWithNotDelete(ctx context.Context, email string) (*User, error)
		FindManyByClassID(ctx context.Context, classID uint64, limit int64) ([]*User, error)
		FindManyByClassIDWithCursor(ctx context.Context, classID, cursor uint64, limit int64) ([]*User, error)
		SwitchUserClass(ctx context.Context, classID uint64, department string, data *User) (sql.Result, error)
		SwitchUserClassWithSession(ctx context.Context, session sqlx.Session, classID uint64, department string, data *User) (sql.Result, error)
	}

	customUserModel struct {
		*defaultUserModel
	}
)

// NewUserModel returns a model for the database table.
func NewUserModel(conn sqlx.SqlConn) UserModel {
	return &customUserModel{
		defaultUserModel: newUserModel(conn),
	}
}

func (m *customUserModel) withSession(session sqlx.Session) UserModel {
	return NewUserModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customUserModel) Insert(ctx context.Context, data *User) (sql.Result, error) {
	//nolint:staticcheck // QF1008: go-zero embedding style retained intentionally
	result, err := m.defaultUserModel.Insert(ctx, data)
	return result, mapDBError(err)
}

func (m *customUserModel) InsertWithSession(ctx context.Context, session sqlx.Session, data *User) (sql.Result, error) {
	return m.withSession(session).Insert(ctx, data)
}

func (m *customUserModel) FindOneWithNotDelete(ctx context.Context, id uint64) (*User, error) {
	query := fmt.Sprintf(`
		select %s from %s
		where id = ? and deleted_at = 0 and status = 1`,
		userRows,
		m.table)

	var resp User
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	return &resp, mapDBError(err)
}

func (m *customUserModel) FindOneByEmailWithNotDelete(ctx context.Context, email string) (*User, error) {
	query := fmt.Sprintf(`
		select %s from %s
		where email = ? and deleted_at = 0`,
		userRows,
		m.table)

	var resp User
	err := m.conn.QueryRowCtx(ctx, &resp, query, email)
	return &resp, mapDBError(err)
}

func (m *customUserModel) FindManyByClassID(ctx context.Context, classID uint64, limit int64) ([]*User, error) {
	query := fmt.Sprintf(`
		select %s from %s
		where class_id = ? and deleted_at = 0 and status = 1
		order by id desc
		limit ?`,
		userRows,
		m.table,
	)

	var resp []*User
	err := m.conn.QueryRowsCtx(ctx, &resp, query, classID, limit)
	return resp, mapDBError(err)
}

func (m *customUserModel) FindManyByClassIDWithCursor(ctx context.Context, classID, cursor uint64, limit int64) ([]*User, error) {
	query := fmt.Sprintf(`
		select %s from %s
		where class_id = ? 
		    and id < ? 
		    and deleted_at = 0 
		    and status = 1
		order by id desc
		limit ?`,
		userRows,
		m.table,
	)

	var resp []*User
	err := m.conn.QueryRowsCtx(ctx, &resp, query, classID, cursor, limit)
	return resp, mapDBError(err)
}

func (m *customUserModel) SwitchUserClass(ctx context.Context, classID uint64, department string, data *User) (sql.Result, error) {
	query := fmt.Sprintf(`
		update %s set class_id = ?,
		              department = ?
		where id = ? 
		    and class_id = ?	
		    and deleted_at = 0 
		    and status = 1`,
		m.table)

	result, err := m.conn.ExecCtx(ctx, query, classID, department, data.Id, data.ClassId)
	return result, mapDBError(err)
}

func (m *customUserModel) SwitchUserClassWithSession(ctx context.Context, session sqlx.Session, classID uint64, department string, data *User) (sql.Result, error) {
	return m.withSession(session).SwitchUserClass(ctx, classID, department, data)
}
