FROM golang:1.23.6-alpine3.21 AS builder

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY . .

RUN go build -o main ./cmd

FROM alpine:latest



WORKDIR /root/

COPY --from=builder /app/main .

COPY .env .

CMD ./main