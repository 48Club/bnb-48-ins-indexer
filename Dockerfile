FROM golang:alpine

WORKDIR /48club
COPY . .
RUN go mod tidy \
    && go build -o app

FROM alpine:3.19

WORKDIR /48club

ENV  TZ=UTC\
    MYSQL_PASSWORD=123456

COPY --from=0 /48club/app /usr/bin/app

RUN apk --no-cache add ca-certificates gcompat tzdata libstdc++

ENTRYPOINT app
