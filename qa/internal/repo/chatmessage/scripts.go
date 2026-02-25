package chatmessage

import (
	_ "embed"
)

//go:embed lua/upsert_message.lua
var upsertMessageLua string
