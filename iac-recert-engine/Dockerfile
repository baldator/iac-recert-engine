FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY ../plugins/ ../plugins/
RUN go build -o ice ./cmd/ice

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/ice .

ENTRYPOINT ["./ice"]
CMD ["run"]
