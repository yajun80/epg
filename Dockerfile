# 构建阶段
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# 交叉编译linux无CGO可执行文件
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o epg-server main.go

# 运行阶段（轻量镜像）
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/epg-server .
# 暴露端口
EXPOSE 8080
# 启动命令
ENTRYPOINT ["./epg-server"]
