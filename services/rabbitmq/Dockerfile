# --- Build Stage ---
FROM golang:1.25.0-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o /rabbitmq-test .

# --- Final Stage ---
# Use a "distroless" image for the final container.
# It contains only the minimum libraries to run the app, making it very secure and small.
FROM gcr.io/distroless/static-debian11

WORKDIR /

# Copy the binary from the builder stage
COPY --from=builder /rabbitmq-test /rabbitmq-test

# Command to run the executable
ENTRYPOINT ["/rabbitmq-test"]
