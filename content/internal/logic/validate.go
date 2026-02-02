package logic

func isValidContentType(tp string) bool {
	return tp == ContentTypeArticle || tp == ContentTypeComic || tp == ContentTypeVideo || tp == ContentTypePodcast
}

func isValidCommentStatus(status uint64) bool {
	return status == CommentStatusNormal || status == CommentStatusCheck || status == CommentStatusShield
}
