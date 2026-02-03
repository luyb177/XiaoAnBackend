package logic

import (
	v1 "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
)

var validContentTypes = map[string]struct{}{
	ContentTypeArticle: {},
	ContentTypeComic:   {},
	ContentTypeVideo:   {},
	ContentTypePodcast: {},
	ContentTypeComment: {},
}

var validCommentStatuses = map[uint64]struct{}{
	CommentStatusNormal: {},
	CommentStatusCheck:  {},
	CommentStatusShield: {},
}

// ValidateContentTypeAndIDRequest
// 接收内容类型和内容ID，如果验证失败则返回一个错误响应，否则返回nil
func ValidateContentTypeAndIDRequest(contentType string, contentID uint64) *v1.Response {
	switch {
	case contentType == "":
		return bad("内容类型不能为空")
	case !isValidContentType(contentType):
		return bad("内容类型不合法")
	case contentID == 0:
		return bad("内容ID不能为0")
	}
	return nil
}

func isValidContentType(tp string) bool {
	_, ok := validContentTypes[tp]
	return ok
}

func isValidCommentStatus(status uint64) bool {
	_, ok := validCommentStatuses[status]
	return ok
}
