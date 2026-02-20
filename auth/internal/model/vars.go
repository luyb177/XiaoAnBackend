package model

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrNotFound       = sqlx.ErrNotFound
	ErrDuplicateEntry = errors.New("duplicate entry")
)
