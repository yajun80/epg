FROM golang:1.21-alpine AS builder

# 设置国内代理
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn

# 安装 git（部分依赖需要拉取 git 仓库）
RUN apk add --no-cache git

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o epg-server main.go

# 构建最小运行镜像
FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/epg-server .

# 暴露服务端口（根据你的实际端口修改）
EXPOSE 8082

# 启动服务
CMD ["./epg-server"]
