package collect

import (
	_ "embed"
)

//go:embed lua/collect.lua
var collectLua string

//go:embed lua/uncollect.lua
var unCollectLua string
