FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .
WORKDIR /app/iac-recert-engine
RUN go mod download
RUN go build -o ice ./cmd/ice

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/iac-recert-engine/ice .

ENTRYPOINT ["./ice"]
CMD ["run"]
