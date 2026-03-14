# 使用官方Go镜像作为构建环境
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装依赖
RUN apk add --no-cache git

# 复制go mod文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o epg-server .

# 使用轻量级镜像作为运行环境
FROM alpine:latest

# 安装CA证书（用于HTTPS请求）
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN adduser -D -g '' appuser

WORKDIR /app

# 从builder阶段复制编译好的二进制文件
COPY --from=builder /app/epg-server .

# 使用非root用户运行
USER appuser

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./epg-server"]