package redisqueue

import (
	_ "embed"
)

//go:embed lua/move_retry_to_pending.lua
var moveRetryToPendingLua string
