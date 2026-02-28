package xtime

import "database/sql"

// UnixOrZero converts sql.NullTime to unix seconds.
// returns 0 if invalid.
func UnixOrZero(nt sql.NullTime) int64 {
	if nt.Valid {
		return nt.Time.Unix()
	}
	return 0
}

// UnixMilliOrZero converts sql.NullTime to unix milliseconds.
func UnixMilliOrZero(nt sql.NullTime) int64 {
	if nt.Valid {
		return nt.Time.UnixMilli()
	}
	return 0
}
