FROM golang:alpine AS builder

# 设置容器工作目录
WORKDIR /go/src/github.com/isgvto/gin-vue-admin-gblog/server
# 注意：在server目录下执行docker build命令
COPY . .

RUN go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.cn,direct \
    && go env -w CGO_ENABLED=0 \
    && go env \
    && go mod tidy \
    && go build -o server .

FROM alpine:latest

LABEL MAINTAINER="lz13526112262@163.com"
# 设置时区
ENV TZ=Asia/Shanghai
RUN apk update && apk add --no-cache tzdata openntpd \
    && ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /go/src/github.com/isgvto/gin-vue-admin-gblog/server

COPY --from=0 /go/src/github.com/isgvto/gin-vue-admin-gblog/server/server ./
COPY --from=0 /go/src/github.com/isgvto/gin-vue-admin-gblog/server/resource ./resource/
COPY --from=0 /go/src/github.com/isgvto/gin-vue-admin-gblog/server/config.docker.yaml ./

# 挂载目录：如果使用了sqlite数据库，容器命令示例：docker run -d -v /宿主机路径/gva.db:/go/src/github.com/isgvto/gin-vue-admin-gblog/server/gva.db -p 8888:8888 --name gva-server-v1 gva-server:1.0
# VOLUME ["/go/src/github.com/isgvto/gin-vue-admin-gblog/server"]

EXPOSE 8888
ENTRYPOINT ./server -c config.docker.yaml
