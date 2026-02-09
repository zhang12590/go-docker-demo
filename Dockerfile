# 构建阶段
FROM golang:1.24.12-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY main.go ./

# 设置构建参数
ARG TARGETOS=linux
ARG TARGETARCH=arm64
ARG VERSION=1.0.0
ARG BUILD_TIME=unknown

# 编译Go程序
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-w -s \
    -X main.Version=${VERSION} \
    -X main.BuildTime=${BUILD_TIME}" \
    -o go-logger .

# 运行阶段 - 使用最小的基础镜像
FROM alpine:latest

# 安装ca-certificates（用于HTTPS请求）
RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

# 创建非root用户运行程序
RUN addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/go-logger /app/go-logger

# 复制时区和证书
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# zoneinfo 文件不在 build 阶段（golang:alpine）镜像中，所以这里无需复制，应直接在运行时安装 tzdata
# （之前 RUN apk add tzdata 时已包含 zoneinfo，Dockerfile 末尾已设置 TZ 环境变量即可）
# 如果要确保 zoneinfo，有时可以用如下命令检查（此处注释掉原 COPY）：
# COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# 设置权限
RUN chown -R appuser:appuser /app
USER appuser

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# 暴露端口
EXPOSE 8080

# 设置环境变量默认值
ENV LOG_INTERVAL=1 \
    LOG_MESSAGE="容器化日志程序运行中" \
    SERVER_PORT=8080 \
    INCLUDE_HTTP=true \
    TZ=UTC

# 启动程序
# 出现 "exec format error" 通常表示你构建的二进制与运行容器的CPU架构不兼容（例如在ARM上用x86_64二进制）。
# 请确保构建阶段（build）和运行阶段（runtime）的架构一致。以下 ENTRYPOINT 本身没有问题，但前提是你构建的是正确架构的目标文件。
ENTRYPOINT ["/app/go-logger"]