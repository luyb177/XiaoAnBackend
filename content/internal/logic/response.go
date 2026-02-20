package logic

import v1 "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

func bad(msg string) *v1.Response {
	return &v1.Response{
		Code:    400,
		Message: msg,
	}
}

func internal(msg string) *v1.Response {
	return &v1.Response{
		Code:    500,
		Message: msg,
	}
}

func notFound(msg string) *v1.Response {
	return &v1.Response{
		Code:    404,
		Message: msg,
	}
}
