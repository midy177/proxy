FROM alpine:3.11

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update && \
    apk add ca-certificates && \
    apk add -U tzdata && \
    rm -rf /var/cache/apk/*

# 拷贝二进制文件
ADD proxy /
ADD conf.yaml /conf.yaml

ENTRYPOINT ["/proxy"]
