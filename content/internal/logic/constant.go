package logic

const (
	ContentTypeArticle = "article"
	ContentTypePodcast = "podcast"
	ContentTypeComic   = "comic"
	ContentTypeVideo   = "video"
	ContentTypeComment = "comment"
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
