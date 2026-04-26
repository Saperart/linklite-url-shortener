FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o shortener ./cmd/shortener

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/shortener ./shortener

EXPOSE 8080

CMD ["./shortener"]