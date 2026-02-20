package constants

const (
	SUPERADMIN = "superadmin"
	CLASSADMIN = "classadmin"
	STUDENT    = "student"
	STAFF      = "staff"
	GUEST      = "guest"
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
	InvalidClassID = iota
)
