FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o quokka-bot /main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /quokka-bot /app/quokka-bot

CMD ["/app/quokka-bot"]