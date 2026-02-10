package logic

const (
	SUPERADMIN = "superadmin"
	CLASSADMIN = "classadmin"
	STUDENT    = "student"
	STAFF      = "staff"
)

const (
	InviteCodeActive   = 1
	InviteCodeInactive = 0
)

const (
	NamePrefix = "小安用户"
)

const (
	// InvalidUserID 不存在的ID
	InvalidUserID = iota
)

const (
	// UserStatusNormal 正常
	UserStatusNormal = iota + 1
	// UserStatusDisable 禁用
	UserStatusDisable
	// UserStatusDeletion 删除
	UserStatusDeletion
)

const (
	// ClassStatusNormal 正常
	ClassStatusNormal = iota + 1
	// ClassStatusFrozen 冻结
	ClassStatusFrozen
	// ClassStatusDismissed 已解散
	ClassStatusDismissed
)
