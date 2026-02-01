package ip2region

import (
	"context"
	"google.golang.org/grpc/metadata"
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
		ClientIP: get("x-client-ip"),
		Country:  get("x-geo-country"),
		Province: get("x-geo-province"),
		City:     get("x-geo-city"),
		ISP:      get("x-geo-isp"),
		ISOCode:  get("x-geo-iso"),
	}

	// 连 IP 都没有，直接认为“无定位信息”
	if loc.ClientIP == "" {
		return nil
	}

	return loc
}
