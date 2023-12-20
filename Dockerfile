FROM golang:1.20-alpine3.19

WORKDIR /48club

RUN apk add --no-cache git \
    && git clone -b main --depth 1 --single-branch https://github.com/48Club/bnb-48-ins-indexer.git /48club \
    && go mod tidy \
    && go build -o app

FROM alpine:3.19

WORKDIR /48club

ENV  TZ=UTC\
    MYSQL_PASSWORD=123456

COPY --from=0 /48club/app /usr/bin/app

RUN apk --no-cache add ca-certificates gcompat tzdata libstdc++

ENTRYPOINT app
