package ip2region

import (
	"context"
	"encoding/base64"

	"google.golang.org/grpc/metadata"
)

const (
	MdKeyClientIP    = "x-client-ip-b64"
	MdKeyGeoCountry  = "x-geo-country-b64"
	MdKeyGeoProvince = "x-geo-province-b64"
	MdKeyGeoCity     = "x-geo-city-b64"
	MdKeyGeoISP      = "x-geo-isp-b64"
	MdKeyGeoISO      = "x-geo-iso-b64"
)

type IPLocation struct {
	ClientIP string
	Country  string
	Province string
	City     string
	ISP      string
	ISOCode  string
}

func GetIPFromMetadata(ctx context.Context) *IPLocation {
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

	loc := &IPLocation{
		ClientIP: decodeValue(get(MdKeyClientIP)),
		Country:  decodeValue(get(MdKeyGeoCountry)),
		Province: decodeValue(get(MdKeyGeoProvince)),
		City:     decodeValue(get(MdKeyGeoCity)),
		ISP:      decodeValue(get(MdKeyGeoISP)),
		ISOCode:  decodeValue(get(MdKeyGeoISO)),
	}

	// 连 IP 都没有，直接认为“无定位信息”
	if loc.ClientIP == "" {
		return nil
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
