FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /metrics ./services/metrics

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /metrics /metrics
EXPOSE 9100
CMD ["/metrics"]
