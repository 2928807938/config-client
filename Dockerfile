# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 复制所有 go.mod 和 go.sum 文件
COPY bom/go.mod bom/go.sum* ./bom/
COPY share/go.mod share/go.sum* ./share/
COPY cmd/api/go.mod cmd/api/go.sum* ./cmd/api/

# 下载依赖（利用 Docker 缓存层）
WORKDIR /app/cmd/api
RUN go mod download

# 返回根目录并复制所有源代码
WORKDIR /app
COPY . .

# 构建应用
WORKDIR /app/cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/cmd/api/main .

# 复制配置文件
COPY config.yaml .

# 设置时区（可选）
ENV TZ=Asia/Shanghai

EXPOSE 8080

CMD ["./main"]
