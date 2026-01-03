# --- Build Stage ---
FROM golang:1.25.0-alpine AS builder

WORKDIR /app
RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-gateway

# --- Runtime Stage ---
FROM gcr.io/distroless/static-debian11

WORKDIR /
COPY --from=builder /api-gateway /api-gateway
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY config.yaml .

EXPOSE 8080
ENTRYPOINT ["/api-gateway"]
