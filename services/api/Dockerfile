FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /api ./services/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /api /api
EXPOSE 8080
CMD ["/api"]
