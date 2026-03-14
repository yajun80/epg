FROM golang:1.21-alpine

WORKDIR /app

# 先安装依赖
RUN go mod init epg 2>/dev/null || true
RUN go get github.com/tidwall/gjson

# 复制代码
COPY main.go .

# 编译
RUN go build -o epg-server

EXPOSE 8080

CMD ["./epg-server"]