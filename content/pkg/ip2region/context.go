package ip2region

import (
	"context"
	"google.golang.org/grpc/metadata"
)

const (
	MdKeyClientIP    = "x-client-ip"
	MdKeyGeoCountry  = "x-geo-country"
	MdKeyGeoProvince = "x-geo-province"
	MdKeyGeoCity     = "x-geo-city"
	MdKeyGeoISP      = "x-geo-isp"
	MdKeyGeoISO      = "x-geo-iso"
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
		ClientIP: get(MdKeyClientIP),
		Country:  get(MdKeyGeoCountry),
		Province: get(MdKeyGeoProvince),
		City:     get(MdKeyGeoCity),
		ISP:      get(MdKeyGeoISP),
		ISOCode:  get(MdKeyGeoISO),
	}

	// 连 IP 都没有，直接认为“无定位信息”
	if loc.ClientIP == "" {
		return nil
	}

	return loc
}
