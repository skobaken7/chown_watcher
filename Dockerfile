FROM golang:1.13-alpine AS builder

ENV GO111MODULE=on
WORKDIR /build

COPY . .
RUN go build .

FROM alpine:latest

RUN apk update && apk add inotify-tools
COPY --from=builder /build/chown_watcher /chown_watcher
CMD /chown_watcher
