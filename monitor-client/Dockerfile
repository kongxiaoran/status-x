# 使用官方的 Golang 镜像作为基础镜像，使用较小的 Alpine 版本
FROM golang:1.21-alpine AS builder

# 设置 Go 模块代理为国内代理
ENV GOPROXY=https://goproxy.cn,direct

# 设置工作目录
WORKDIR /app

# 将当前目录中的所有文件复制到 /app 目录下
COPY . .

# 拉取依赖包并编译程序
RUN go mod tidy && go build -o /app/monitor .

# 使用一个更小的基础镜像来运行程序，减小镜像大小
FROM alpine:3.18

# 设置工作目录
WORKDIR /root/

# 从 builder 镜像复制生成的二进制文件到当前镜像
COPY --from=builder /app/monitor .

# 允许外部设置环境变量
ENV INFLUX_TOKEN="" \
    INFLUX_URL="localhost:8086" \
    SERVER_URL="localhost:8080" \
    COLLECT_FREQUENCY="1"

# 启动程序
CMD ["./monitor"]
