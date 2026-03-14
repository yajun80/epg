FROM golang:1.21-alpine AS builder

WORKDIR /app

# 先复制依赖文件
COPY go.mod ./
# 先不复制 go.sum，让 go mod 生成

# 下载依赖
RUN go mod download

# 复制源代码
COPY main.go .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -o epg-server

# 最终镜像
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/epg-server .

EXPOSE 8080

CMD ["./epg-server"]