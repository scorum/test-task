FROM golang:1.20.2-alpine3.17 as builder

WORKDIR /app
COPY . .

RUN go build -o "bin/api" cmd/api/main.go

FROM alpine:3.17

RUN apk update && apk add ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/api /app/api
COPY --from=builder /app/configs/config.yml /app/configs/config.yml
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080

CMD ["/app/api"]