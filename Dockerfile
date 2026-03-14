# 使用官方Go镜像作为构建环境
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装依赖
RUN apk add --no-cache git ca-certificates

# 复制go mod文件（如果存在）
COPY go.mod go.sum* ./
RUN if [ -f go.mod ]; then \
        go mod download; \
    fi

# 复制源代码
COPY *.go ./

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o epg-server .

# 使用轻量级镜像作为运行环境
FROM alpine:latest

# 安装CA证书（用于HTTPS请求）和时区数据
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为上海时间
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN adduser -D -g '' appuser

WORKDIR /app

# 从builder阶段复制编译好的二进制文件
COPY --from=builder /app/epg-server .

# 创建缓存目录
RUN mkdir -p /app/cache && chown -R appuser:appuser /app

# 使用非root用户运行
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q --spider http://localhost:8080/epg || exit 1

# 运行应用
CMD ["./epg-server"]