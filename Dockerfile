FROM golang:rc-alpine3.12 as builder

WORKDIR /buildtmp
COPY . /buildtmp
ENV GOOS linux
ENV CGO_ENABLED 0
ENV GO111MODULE: on
ENV GOPROXY https://goproxy.io
RUN go build -a -installsuffix cgo -ldflags '-w -s'

FROM alpine:3.12

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update && \
    apk add ca-certificates && \
    apk add -U tzdata && \
    rm -rf /var/cache/apk/*

# 拷贝二进制文件
COPY --from=builder /buildtmp/proxy /
COPY conf.yaml /conf.yaml

ENTRYPOINT ["/proxy"]
