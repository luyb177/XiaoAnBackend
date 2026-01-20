package redisqueue

import (
	_ "embed"
)

//go:embed lua/move_retry_to_pending.lua
var moveRetryToPendingLua string

//go:embed lua/move_processing_to_dlq.lua
var moveProcessingToDLQLua string

//go:embed lua/move_processing_to_retry.lua
var moveProcessingToRetryLua string
