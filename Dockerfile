# Dockerfile
FROM golang:1.20 as builder

WORKDIR /app

COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/promo-cmd

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 2112

CMD ["./main"]