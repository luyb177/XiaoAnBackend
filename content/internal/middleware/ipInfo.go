package middleware

import "context"

type IPInfo struct {
	ClientIP string
	Country  string
	Province string
	City     string
	ISP      string
	ISOCode  string
}

func GetIPInfo(ctx context.Context) (*IPInfo, bool) {
	info, ok := ctx.Value(ctxKeyIPInfo).(*IPInfo)
	return info, ok
}

func DefaultIPInfo() *IPInfo {
	return &IPInfo{
		ClientIP: "",
		Country:  "未知",
		Province: "未知",
		City:     "未知",
		ISP:      "未知",
		ISOCode:  "未知",
	}
}
