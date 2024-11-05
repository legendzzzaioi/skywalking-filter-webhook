FROM swr.cn-east-3.myhuaweicloud.com/woody-public/golang:1.23.2-alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOPROXY https://goproxy.cn,direct
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /build

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o /app/skywalking-filter-webhook .

FROM swr.cn-east-3.myhuaweicloud.com/woody-public/alpine:3.20.3
ENV TZ Asia/Shanghai
WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai

COPY --from=builder /app/skywalking-filter-webhook /app/skywalking-filter-webhook
COPY --from=builder /build/config.yaml /app/config.yaml

EXPOSE 8000/tcp

CMD ["./skywalking-filter-webhook", "-f", "config.yaml"]