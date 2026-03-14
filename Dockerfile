FROM golang:1.21-alpine

WORKDIR /app

# 复制所有文件
COPY . .

# 下载依赖并编译
RUN go mod tidy && go build -o epg-server

EXPOSE 8080

CMD ["./epg-server"]