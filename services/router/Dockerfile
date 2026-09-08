FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /router ./services/router

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /router /router
EXPOSE 8081
CMD ["/router"]
