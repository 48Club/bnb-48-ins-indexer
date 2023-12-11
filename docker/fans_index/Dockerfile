FROM golang:1.20 as builder

ENV GO111MODULE=on

WORKDIR /www/fans_index
COPY . .
RUN go mod tidy

RUN go build -o fans_index main.go

FROM ubuntu:22.04

WORKDIR /www/fans_index
ENV TIME_ZONE=UTC
RUN ln -snf /usr/share/zoneinfo/$TIME_ZONE /etc/localtime && echo $TIME_ZONE > /etc/timezone
RUN apt-get update && apt-get install -y ca-certificates

COPY --from=builder /www/fans_index/fans_index /www/fans_index

ENTRYPOINT ["./fans_index", "fans_index"]
