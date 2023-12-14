FROM golang:1.20 as builder

ENV GO111MODULE=on

WORKDIR /www/api
COPY . .
RUN go mod tidy

RUN go build -o api main.go

FROM ubuntu:22.04

WORKDIR /www/api
ENV TIME_ZONE=UTC
RUN ln -snf /usr/share/zoneinfo/$TIME_ZONE /etc/localtime && echo $TIME_ZONE > /etc/timezone
RUN apt-get update && apt-get install -y ca-certificates

COPY --from=builder /www/api/api /www/api

ENTRYPOINT ["./api", "api"]
