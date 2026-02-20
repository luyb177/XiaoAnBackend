package logic

import v1 "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"

func bad(msg string) *v1.Response {
	return &v1.Response{
		Code:    400,
		Message: msg,
		Data:    nil,
	}
}

func internal(msg string) *v1.Response {
	return &v1.Response{
		Code:    500,
		Message: msg,
		Data:    nil,
	}
}

func notFound(msg string) *v1.Response {
	return &v1.Response{
		Code:    404,
		Message: msg,
	}
}
