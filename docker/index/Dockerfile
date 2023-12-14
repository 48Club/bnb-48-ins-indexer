FROM golang:1.20 as builder

ENV GO111MODULE=on

WORKDIR /www/index
COPY . .
RUN go mod tidy

RUN go build -o index main.go

FROM ubuntu:22.04

WORKDIR /www/bnb48_index
ENV TIME_ZONE=UTC
RUN ln -snf /usr/share/zoneinfo/$TIME_ZONE /etc/localtime && echo $TIME_ZONE > /etc/timezone
RUN apt-get update && apt-get install -y ca-certificates

COPY --from=builder /www/index/index /www/bnb48_index

ENTRYPOINT ["./index", "index"]
