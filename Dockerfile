FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o tldr-telegram-bot ./cmd/bot/main.go

FROM ubuntu:latest

RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata    

WORKDIR /root/

COPY --from=builder /app/tldr-telegram-bot .

CMD ["./tldr-telegram-bot"]
