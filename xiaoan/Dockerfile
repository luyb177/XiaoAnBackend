FROM golang:1.25.3-alpine AS builder

ENV CGO_ENABLED=0
RUN apk add --no-cache tzdata
ENV GOPROXY=https://goproxy.cn,direct

# 关键：工作目录指向 xiaoan
WORKDIR /build/xiaoan

# 先拷贝整个仓库（为了支持 ../auth、../qa）
COPY . /build

# 下载依赖（此时 go.mod 在当前目录）
RUN go mod download

# 构建
RUN go build -ldflags="-s -w" -o /app/xiaoan xiaoan.go



FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai

ENV TZ=Asia/Shanghai
WORKDIR /app

COPY --from=builder /app/xiaoan /app/xiaoan

# 配置文件目录（可被 volume 覆盖）
COPY xiaoan/etc /app/etc

CMD ["./xiaoan", "-f", "etc/xiaoan.yaml"]
