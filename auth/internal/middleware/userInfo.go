package middleware

import "context"

type UserInfo struct {
	UID    uint64
	Role   string
	Status int64
}

func MustGetUser(ctx context.Context) *UserInfo {
	uid, _ := ctx.Value(ctxKeyUserID).(uint64)
	role, _ := ctx.Value(ctxKeyUserRole).(string)
	status, _ := ctx.Value(ctxKeyUserStatus).(int64)

	return &UserInfo{
		UID:    uid,
		Role:   role,
		Status: status,
	}
}
