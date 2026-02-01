package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"

	contentIp2region "github.com/luyb177/XiaoAnBackend/content/pkg/ip2region"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/config"

	"github.com/lionsoul2014/ip2region/binding/golang/service"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type IPMiddleware struct {
	logx.Logger
	ip2region *service.Ip2Region
}

func NewIPMiddleware(cfg config.IP2RegionConfig) *IPMiddleware {
	l := logx.WithContext(context.Background())
	// 1. 创建 v4 配置
	v4Config, err := service.NewV4Config(service.VIndexCache, cfg.V4, 20)
	if err != nil {
		logx.Must(err)
	}

	// 2. 创建 v6 配置
	v6Config, err := service.NewV6Config(service.VIndexCache, cfg.V6, 20)
	if err != nil {
		logx.Must(err)
	}

	// 3. 创建 ip2region 实例
	ip2region, err := service.NewIp2Region(v4Config, v6Config)
	if err != nil {
		logx.Must(err)
	}

	return &IPMiddleware{
		Logger:    l,
		ip2region: ip2region,
	}
}

func (m *IPMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 获取客户端 IP 地址
		ip := getClientIP(r)

		var ipLocation *IPLocation
		if ip != "" && m.ip2region != nil {
			region, err := m.ip2region.SearchByStr(ip)
			if err != nil {
				m.Errorf("ip2region search err: %v", err)
			} else {
				ipLocation = parseIPRegion(region)
			}
		}

		// 2. 将 IP 地址和地理位置存储在 metadata 中
		ctx := r.Context()

		if ip != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, contentIp2region.MdKeyClientIP, ip)
		}
		if ipLocation != nil {
			ctx = metadata.AppendToOutgoingContext(
				ctx,
				contentIp2region.MdKeyGeoCountry, ipLocation.Country,
				contentIp2region.MdKeyGeoProvince, ipLocation.Province,
				contentIp2region.MdKeyGeoCity, ipLocation.City,
				contentIp2region.MdKeyGeoISP, ipLocation.ISP,
				contentIp2region.MdKeyGeoISO, ipLocation.ISOCode,
			)
		}

		next(w, r.WithContext(ctx))
	}
}

func getClientIP(r *http.Request) string {
	// 1. X-Forwarded-For
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// client, proxy1, proxy2
		ips := strings.Split(xff, ",")
		ip := strings.TrimSpace(ips[0])
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	// 2. X-Real-IP
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		if net.ParseIP(xrip) != nil {
			return xrip
		}
	}

	// 3. RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && net.ParseIP(host) != nil {
		return host
	}

	return ""
}

type IPLocation struct {
	Country  string
	Province string
	City     string
	ISP      string
	ISOCode  string
}

func parseIPRegion(region string) *IPLocation {
	// region 格式：国家|省份|城市|ISP|iso-alpha2-code(国家两字母简称)
	parts := strings.Split(region, "|")
	if len(parts) < 5 {
		return nil
	}

	return &IPLocation{
		Country:  parts[0],
		Province: parts[1],
		City:     parts[2],
		ISP:      parts[3],
		ISOCode:  parts[4],
	}
}
