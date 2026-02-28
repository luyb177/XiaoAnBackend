package openai

import (
	_ "embed"
)

//go:embed prompts/system.txt
var SystemPrompt string

//go:embed prompts/title.txt
var TitlePrompt string
