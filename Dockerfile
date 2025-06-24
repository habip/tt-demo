FROM golang:1.23.5-alpine3.21 AS builder

RUN go version

COPY . /tt-demo/
WORKDIR /tt-demo/

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

RUN go mod download
RUN go build -o ./.bin/tt-demo -tags=go_tarantool_ssl_disable ./cmd/app/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /tt-demo/.bin/tt-demo .
COPY --from=builder /tt-demo/config/local.yaml config/local.yaml
RUN touch .env

EXPOSE 8000

CMD /app/tt-demo