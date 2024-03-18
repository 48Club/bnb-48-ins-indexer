FROM golang:1.20-alpine3.19

WORKDIR /48club

COPY . .
RUN go mod tidy \
    && cd bot \
    && go build -o bot

FROM alpine:3.19

ENV  TZ=UTC\
    TOKEN=123456

COPY --from=0 /48club/bot/bot /usr/bin/bot

RUN apk --no-cache add ca-certificates gcompat tzdata libstdc++

ENTRYPOINT bot
