package tasks

type TaskPrefix string

const (
	ChatSessionTaskPrefix TaskPrefix = "chat_session_task"
	ChatMessageTaskPrefix TaskPrefix = "chat_message_task"
)
