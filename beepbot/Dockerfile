# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git

# 复制 go.mod 和 go.sum 以利用缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o beepbot-api ./cmd/api

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 从构建阶段复制二进制文件
COPY --from=builder /app/beepbot-api .

# 创建数据目录
RUN mkdir -p /data/beepbot/skills

# 暴露端口
EXPOSE 8888

# 运行应用
ENTRYPOINT ["./beepbot-api"]
CMD ["-config", "/app/config.json"]