package logic

// 是否有消息
const (
	HasMessageNo  = iota // 没有消息
	HasMessageYes        // 有消息
)

// 会话状态
const (
	SessionStatusEmpty    = iota // 空会话（未开始）
	SessionStatusRunning         // 进行中
	SessionStatusFinished        // 已结束
)

// 是否置顶
const (
	PinnedNo = iota
	PinnedYes
)

// 关系状态
const (
	RelationStatusNormal = iota
	RelationStatusPending
)

const (
	MessageRoleUser      = 0
	MessageRoleAssistant = 1
	MessageRoleSystem    = 2
)

const (
	MessageTypeText  = 0
	MessageTypeImage = 1
	MessageTypeFile  = 2
)

const (
	MessageStatusGenerating = 0
	MessageStatusSuccess    = 1
	MessageStatusFailed     = 2
)

const (
	FinishReasonStop  = "stop"
	FinishReasonError = "error"
)

const (
	DefaultPageSize = 10
	MaxPageSize     = 50
)
