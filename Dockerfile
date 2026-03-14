FROM golang:1.21-alpine

WORKDIR /app

# 复制代码
COPY main.go .

# 初始化模块并安装依赖
RUN go mod init epg && \
    go get github.com/tidwall/gjson && \
    go build -o epg-server

EXPOSE 8080

CMD ["./epg-server"]