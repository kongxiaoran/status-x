# 使用官方的 Golang 镜像作为基础镜像，使用较小的 Alpine 版本
FROM golang:1.21-alpine AS builder

# 设置 Go 模块代理为国内代理
ENV GOPROXY=https://goproxy.cn,direct

# 设置工作目录
WORKDIR /app

# 复制当前目录下的所有文件到容器的 /app 目录下
COPY . .

# 构建 Go 应用
RUN go build -o main .

# 使用 Alpine Linux 的官方镜像作为最终的运行环境
FROM alpine:3.18

# 设置工作目录
WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache ca-certificates tzdata

# 复制构建好的二进制文件到目标镜像
COPY --from=builder /app/main /app

# 复制前端模板文件
COPY --from=builder /app/templates /app/templates

#复制配置文件
COPY --from=builder /app/kubeconfig.yaml /app/kubeconfig.yaml

# 暴露容器需要监听的端口
EXPOSE 12800

# 启动命令
CMD ["./main"]
