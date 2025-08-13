FROM golang:1.24-alpine AS builder
RUN apk add --no-cache gcc musl-dev git
WORKDIR /app

ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.google.cn

COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./

RUN CGO_ENABLED=1 GOOS=linux go build -o /app/dash ./main.go

FROM alpine:3.22
WORKDIR /app

COPY --from=builder /app/dash /app/dash
# 复制配置文件到容器
COPY conf /app/conf
# 复制静态资源到容器
COPY resource/static /app/resource/static

# 暴露服务端口 8080（与应用配置一致）
EXPOSE 8080

# 指定容器启动命令
CMD ["/app/dash"]