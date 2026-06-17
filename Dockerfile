# 第一阶段：编译
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 复制依赖文件并下载
COPY go.mod go.sum ./

RUN go env -w GOPROXY=https://goproxy.cn,direct && \
	go env -w GOSUMDB=sum.golang.google.cn

RUN go mod download

# 复制源代码并编译
COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o shortlink ./cmd/server

# 第二阶段：运行
FROM alpine:latest

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/shortlink ./

# 复制配置文件（使用默认配置）
COPY ./configs/config.toml ./configs/

EXPOSE 8080

CMD ["./shortlink"]