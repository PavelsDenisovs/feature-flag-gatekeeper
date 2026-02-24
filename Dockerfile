FROM golang:1.26-alpine3.23 AS builder

WORKDIR /app
COPY . .

RUN go build -o ffg ./cmd/ffg

FROM alpine:3.23

WORKDIR /app
COPY --from=builder /app/ffg .

CMD ["./ffg", "serve"]