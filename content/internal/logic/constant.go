package logic

const (
	SUPERADMIN = "superadmin"
	CLASSADMIN = "classadmin"
	STUDENT    = "student"
	STAFF      = "staff"
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
	ArticleContentImage = iota + 1
)

var (
	ArticleImageMap = map[int64]struct{}{
		ArticleContentImage: {}, // 内容
	}
)
