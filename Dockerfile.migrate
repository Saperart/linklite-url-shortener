FROM golang:1.25-alpine AS builder

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM alpine:3.22

COPY --from=builder /go/bin/goose /usr/local/bin/goose

WORKDIR /app

COPY migrations /migrations

ENTRYPOINT ["goose"]