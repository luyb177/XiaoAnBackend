package logic

import "github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

func BadResponse(msg string) *types.Response {
	return &types.Response{
		Code:    400,
		Message: msg,
		Data:    &types.EmptyResponse{},
	}
}

func InternalErrorResponse(msg string) *types.Response {
	return &types.Response{
		Code:    500,
		Message: msg,
		Data:    &types.EmptyResponse{},
	}
}
