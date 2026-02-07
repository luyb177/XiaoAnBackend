package middleware

import (
	"context"
)

type UserInfo struct {
	UID    uint64
	Role   string
	Status int64
}

func GetUser(ctx context.Context) (*UserInfo, bool) {
	user, ok := ctx.Value(ctxKeyUser).(*UserInfo)
	return user, ok
}
