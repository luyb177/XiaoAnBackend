package logic

const (
	SUPERADMIN = "superadmin"
	CLASSADMIN = "classadmin"
	STUDENT    = "student"
	STAFF      = "staff"
)

const (
	ContentTypeArticle = "article"
	ContentTypePodcast = "podcast"
	ContentTypeComic   = "comic"
	ContentTypeVideo   = "video"
	ContentTypeComment = "comment"
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
	RelationStatusNormal = iota
	RelationStatusPending
)

const (
	PodcastStatusPublished = iota
	PodcastStatusDraft
)

const (
	ComicStatusPublished = iota
	ComicStatusDraft
)

const (
	CommentStatusNormal = iota
	CommentStatusCheck
	CommentStatusShield
)

const (
	RootCommentParentID = 0
)
