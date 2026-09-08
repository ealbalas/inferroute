FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /worker ./services/worker

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /worker /worker
EXPOSE 9001
CMD ["/worker"]
