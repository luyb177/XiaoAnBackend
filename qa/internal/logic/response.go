package logic

import v1 "github.com/luyb177/XiaoAnBackend/qa/pb/qa/v1"

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