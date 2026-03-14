# 多阶段构建：减小镜像体积
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# 编译 Linux amd64 架构的二进制文件
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o epg-server main.go

# 运行阶段
FROM alpine:3.19
WORKDIR /app
# 从构建阶段复制编译产物
COPY --from=builder /app/epg-server .

EXPOSE 8080
# 使用非 root 用户运行 (安全最佳实践)
RUN adduser -D appuser
USER appuser

ENTRYPOINT ["./epg-server"]
