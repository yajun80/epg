FROM golang:1.21-alpine

WORKDIR /app

# 复制依赖文件
COPY go.mod ./
RUN go mod download

# 复制源码
COPY main.go .

# 编译
RUN go build -o epg-server

# 暴露端口
EXPOSE 8080

# 运行
CMD ["./epg-server"]
