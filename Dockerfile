FROM golang:1.20rc2-alpine3.17 AS buider

COPY . /telegram-bot-pocket
WORKDIR /telegram-bot-pocket

RUN go mod download
RUN go build -o ./bin/bot/main cmd/bot/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=buider /telegram-bot-pocket/bin/bot .
COPY --from=buider /telegram-bot-pocket/cmd/bot/configs configs/

EXPOSE 8080

CMD ["./main"]