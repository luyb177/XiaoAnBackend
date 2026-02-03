package like

import (
	_ "embed"
)

//go:embed lua/like.lua
var likeLua string

//go:embed lua/unlike.lua
var unlikeLua string
