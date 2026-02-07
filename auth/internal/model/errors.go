package model

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func mapDBError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sqlx.ErrNotFound) {
		return ErrNotFound
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1062:
			return ErrDuplicateEntry
		}
	}

	return err
}
