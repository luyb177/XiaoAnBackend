package middleware

import (
	"context"
	"encoding/base64"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const ctxKeyIPInfo ctxKey = "ip_info"

const (
	MdKeyClientIP    = "x-client-ip-b64"
	MdKeyGeoCountry  = "x-geo-country-b64"
	MdKeyGeoProvince = "x-geo-province-b64"
	MdKeyGeoCity     = "x-geo-city-b64"
	MdKeyGeoISP      = "x-geo-isp-b64"
	MdKeyGeoISO      = "x-geo-iso-b64"
)

// 需要记录 IP 信息的方法
//
//	/<proto包名>.<ServiceName>/<MethodName>
var ipRecordMethods = map[string]struct{}{
	"/content.ContentService/AddComment": {},
}

// IPUnaryInterceptor IP 信息拦截器
func IPUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 1. 跳过不需要记录 IP 的接口
	if _, ok := ipRecordMethods[info.FullMethod]; !ok {
		return handler(ctx, req)
	}

	// 2. 需要记录 IP 的接口，从 metadata 取 IP 信息
	ipLocation := extractIPFromMetadata(ctx)

	// 3. 写入 context，供 logic 使用
	if ipLocation != nil {
		ctx = context.WithValue(ctx, ctxKeyIPInfo, ipLocation)
	}
	return handler(ctx, req)
}

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

func extractIPFromMetadata(ctx context.Context) *IPInfo {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}

	get := func(key string) string {
		if v := md.Get(key); len(v) > 0 {
			return v[0]
		}
		return ""
	}

	loc := &IPInfo{
		ClientIP: decodeValue(get(MdKeyClientIP)),
		Country:  decodeValue(get(MdKeyGeoCountry)),
		Province: decodeValue(get(MdKeyGeoProvince)),
		City:     decodeValue(get(MdKeyGeoCity)),
		ISP:      decodeValue(get(MdKeyGeoISP)),
		ISOCode:  decodeValue(get(MdKeyGeoISO)),
	}

	// 连 IP 都没有，直接认为“无定位信息”
	if loc.ClientIP == "" {
		return DefaultIPInfo()
	}

	return loc
}

// decodeValue 对 base64 编码的字符串进行解码，解码失败则返回空字符串
func decodeValue(s string) string {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return ""
	}
	return string(b)
}
