FROM golang:1.25.3-alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED=0
ENV GOPROXY=https://goproxy.cn,direct

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o /app/qa qa.go


FROM scratch

# 证书 & 时区
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/qa /app/qa

# ⚠️ 不再 COPY etc / data
# 配置 & 数据全部靠 volume

CMD ["./qa", "-f", "etc/qa.yaml"]
